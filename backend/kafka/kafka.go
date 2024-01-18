package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/benjamonnguyen/opendoorchat/backend"
	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

type KafkaConsumerClient interface {
	SetRecordHandler(string, func(*kgo.Record)) error
	Poll(context.Context)
	Shutdown()
}

func NewSplitConsumerClient(
	ctx context.Context,
	cfg backend.KafkaConfig,
	groupId string,
) *splitConsumerClient {
	s := &splitConsumerClient{
		consumers:      make(map[tp]*pConsumer),
		recordHandlers: make(map[string]func(*kgo.Record)),
	}

	tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}}
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(cfg.Brokers, ",")...),
		kgo.SASL(scram.Sha512(func(context.Context) (scram.Auth, error) {
			return scram.Auth{
				User: cfg.User,
				Pass: cfg.Password,
			}, nil
		})),
		kgo.Dialer(tlsDialer.DialContext),
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevel(cfg.LogLevel), nil)),
		kgo.DisableIdempotentWrite(),
		kgo.OnPartitionsAssigned(s.assigned),
		kgo.OnPartitionsRevoked(s.revoked),
		kgo.OnPartitionsLost(s.lost),
		kgo.ConsumerGroup(groupId),
		kgo.AutoCommitMarks(),
		kgo.BlockRebalanceOnPoll(),
	}
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed creating kafka client")
	}
	if err = cl.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed pinging kafka client")
	}

	s.cl = cl

	return s
}

func (s *splitConsumerClient) SetRecordHandler(
	topic string,
	recordHandler func(*kgo.Record),
) error {
	if topic == "" {
		return errors.New("missing required topic")
	}
	s.recordHandlers[topic] = recordHandler
	s.cl.AddConsumeTopics(topic)
	return nil
}

func (s *splitConsumerClient) Poll(ctx context.Context) {
	for {
		// PollRecords is strongly recommended when using
		// BlockRebalanceOnPoll. You can tune how many records to
		// process at once (upper bound -- could all be on one
		// partition), ensuring that your processor loops complete fast
		// enough to not block a rebalance too long.
		fetches := s.cl.PollRecords(ctx, 10000)
		if fetches.IsClientClosed() {
			return
		}
		if fetches.IsClientClosed() {
			log.Info().Msg("kafka client closed")
			return
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			// Note: you can delete this block, which will result
			// in these errors being sent to the partition
			// consumers, and then you can handle the errors there.
			log.Error().
				Interface("errors", errs).
				Msg("failed PollFetches")
			return
		}
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			tp := tp{p.Topic, p.Partition}

			// Since we are using BlockRebalanceOnPoll, we can be
			// sure this partition consumer exists:
			//
			// * onAssigned is guaranteed to be called before we
			// fetch offsets for newly added partitions
			//
			// * onRevoked waits for partition consumers to quit
			// and be deleted before re-allowing polling.
			s.consumers[tp].recs <- p
		})
		s.cl.AllowRebalance()
	}
}

func (s *splitConsumerClient) Shutdown() {
	s.cl.CloseAllowingRebalance()
}

type splitConsumerClient struct {
	consumers      map[tp]*pConsumer
	recordHandlers map[string]func(*kgo.Record)
	cl             *kgo.Client
}

type tp struct {
	t string
	p int32
}

type pConsumer struct {
	cl        *kgo.Client
	topic     string
	partition int32

	recordHandler func(*kgo.Record)

	quit chan struct{}
	done chan struct{}
	recs chan kgo.FetchTopicPartition
}

func (pc *pConsumer) consume() {
	defer close(pc.done)
	log.Info().
		Str("topic", pc.topic).
		Int32("partition", pc.partition).
		Msg("starting split consumer")
	defer log.Info().
		Str("topic", pc.topic).
		Int32("partition", pc.partition).
		Msg("killing split consumer")
	for {
		select {
		case <-pc.quit:
			return
		case p := <-pc.recs:
			p.EachRecord(pc.recordHandler)
			pc.cl.MarkCommitRecords(p.Records...)
		}
	}
}

func (s *splitConsumerClient) assigned(
	_ context.Context,
	cl *kgo.Client,
	assigned map[string][]int32,
) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {
			pc := &pConsumer{
				cl:        cl,
				topic:     topic,
				partition: partition,

				recordHandler: s.recordHandlers[topic],

				quit: make(chan struct{}),
				done: make(chan struct{}),
				recs: make(chan kgo.FetchTopicPartition, 5),
			}
			s.consumers[tp{topic, partition}] = pc
			go pc.consume()
		}
	}
}

func (s *splitConsumerClient) revoked(
	ctx context.Context,
	cl *kgo.Client,
	revoked map[string][]int32,
) {
	s.killConsumers(revoked)
	if err := cl.CommitMarkedOffsets(ctx); err != nil {
		log.Error().Err(err).Msg("failed revoke commit")
	}
}

func (s *splitConsumerClient) lost(_ context.Context, cl *kgo.Client, lost map[string][]int32) {
	s.killConsumers(lost)
	// Losing means we cannot commit: an error happened.
}

func (s *splitConsumerClient) killConsumers(lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, partition := range partitions {
			tp := tp{topic, partition}
			pc := s.consumers[tp]
			delete(s.consumers, tp)
			close(pc.quit)
			wg.Add(1)
			go func() { <-pc.done; wg.Done() }()
		}
	}
}
