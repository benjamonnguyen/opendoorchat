package consumer

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/model"
	usermodel "github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"github.com/jhillyerd/enmime"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	mailerMsgId = "mailerMsgId"
	fromEmail   = "johnsmith@yahoo.com"
	toUser      = "ben"
)

var (
	emailSvc *emailService
	tMailer  *testMailer
	sender   = usermodel.User{
		FirstName: "John",
		LastName:  "Smith",
		Email:     fromEmail,
		Username:  "johnsmith",
	}
	rcpt = usermodel.User{
		FirstName: "Ben",
		LastName:  "N",
		Email:     "ben@yahoo.com",
		Username:  "ben",
	}
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

func TestForwardEmail(t *testing.T) {
	emailSvc = new(emailService)
	tMailer = new(testMailer)

	const (
		emailData = "Received: from sonic313-56.consmr.mail.ne1.yahoo.com (sonic313-56.consmr.mail.ne1.yahoo.com [66.163.185.31])\r\n\tby benjamins-air.lan (Haraka/3.0.2) with ESMTP id 310BBEB2-8575-40CE-BAB5-DD7176D59EC5.1\r\n\tenvelope-from <johnsmith@yahoo.com>;\r\n\tFri, 10 Nov 2023 01:11:13 -0800\r\nDKIM-Signature: v=1; a=rsa-sha256; c=relaxed/relaxed; d=yahoo.com; s=s2048; t=1699607468; bh=O+eQOZb0WApF01OBl7YfH5Bc4Yo1hLik9FBKxwjYmIE=; h=From:Subject:Date:References:In-Reply-To:To:From:Subject:Reply-To; b=fBFTz+0eqhmsoYyW9z3qPbE0PVQsFRqfptWMNrkcCemkzCUuQZo6qDBtPxeBHsn2jxWzsDWO9nTPz7hPwYzZAoo1ocVtgsMVff82165Aeah5xQYESMHqq+lkFZqaZhxWAISn995qy9aGxtEXJGNJELnQNvJFfWzCngtVN8xKcKun0Z+uGmqBqcnxXf7lQI0Csu9IJ54jT1rK5KTTslsOQRhKzg39uCC4KePfF3FeLkzzOa4hrCVJb3As50OJzcschgIjlpNWjwNcZkpLTZVreR5YUae6e3kl4fAqmbS/mgzzA49y0E1JZhwMc6GCgT3nh2FLg6e+aPcNNhLtrYnymg==\r\nX-SONIC-DKIM-SIGN: v=1; a=rsa-sha256; c=relaxed/relaxed; d=yahoo.com; s=s2048; t=1699607468; bh=bw9L71hs8r0+l+uKe9AjTxPJVQbYQNpcj7j2O2zCu52=; h=X-Sonic-MF:From:Subject:Date:To:From:Subject; b=ZivJ3WzdQ3bQDwpZUc2ZRpRmMK+4fYS6J60PUUuvyImsj7zny6RQuQisnxeFTiNZ4f6svfBWD+6/GtIc+tAigcSm879Ex18yfstMVd/RHHrts3pU5d3FJLutVWv9lSBPGNcZ5ARLeGiOntVwsJGGOZ6OWADTYErlBKwQonZwtv8y6+z7VWPtqPqrt7AICUe+LqLKmulxxa/675oQWxgZCVG1GoDecD6F1tTEmPylgInpXzEzCn5YvyDrYG71IozXnydXgXN7MCY8zZ9D3ODg3CtFr81KvX/MI+/uHbl4WMDp3QbSoQq4ePBZjGQH8CrRhekHLoD8fhcIPRHGyRrJEA==\r\nX-YMail-OSG: B8guXgQVM1lPJUPiDe_1qyTsUe03cODHi3Jx3TXAeAQ373GEXVyIPxWwHWMWgW0\r\n qZUp2YBQN94ghq57iirAQQVYB.DMMQkSe1DVflL.ev3VoS1auQ8QxTwpo61C.CBtQuhRRPZ8QX1O\r\n JZ6RKta3.Pld2dOAFCna9D41Q_oEYVvbJY9mOx8KxfWu.N8lSOE5.O_G3bRJacOMETXDfK.1khSh\r\n UmocUl0R5YVCdqRhU1fuyAWQcSxWsMJfANu1lsoih.YA5JX0LGefb5L2sRCLedBUI_VFHNExmN1A\r\n c0YEs98Q238hQskvJyDZlZuQ3CFtjAn_IQpZPTyVd6nEA5XQyQejqm9RzUMJlU8zqnRkMT23m1IH\r\n jqOnTUeS0cTTYOVFqrP3lfc0icQCqyca_fWN2vf8yFA9T_wHyoyyb9co0xDgK5YLFP1qlGtSg9SA\r\n OW6G2BcNbnE2JWQJPhYf0z4NCt8QOPiJax8O3vEwzz8LQYPZrJaCFLQWSIiqnxWoIB6VtFXMYBub\r\n QOFlbBPXcgrqdTrdf_xwSTcrZOOQVe2qxfIcKFUcC5BNqJDzPIM_yxRkfFM78Emft7L9xYILqgV2\r\n 7W_CwW67F_ZvzSQAdrN9KIxx7bKqIT_b3d2d8t6IYb2gTLERX3s_fb9Q3YDCTugpmV7G3jyMI3ej\r\n IY0gd4Bty4z3oqsY3Yq5WltDWfzhvMod2dE14TcpEdZn64X2PuLpvDD7jJjNqZl18irr767tO.tw\r\n ks27D.tdeC9GV0dHzyywbMFrKJHjyMKaoJLyAYP_AYUVbpSuo0O.82cd5DdSta2Xzgs7hyMMyzyX\r\n mdz3CEzOWcJB2ON8gBWmhidHfmJwbKyEFXkBhx1WzJYIMJzBgF07lT2.1_.idSe.QTgBTINN1e9n\r\n FQAputbipHyHkIhQDIdCvEOZ5cJST6w974joAVnR8UmvR0ynchfAzwrbV1ix6FGKI8VnS6rvMhYx\r\n XyiqJ5JXYSMlRrdWJpBsBlnQVDRe4Y3spbL2DIlGlgtd0qciMvdQFrYbs6ykekowvoctg5MY2hkg\r\n eBs13SFPaeFPKmmPOga5daOjsDB_GiTNWpc19s1ra3fIAwhLM0_oBMDEILelGiSQcggV0E_cr0Yd\r\n jbnIkxm_YGjgiOb5xj3gu3acC0CzfPnlGgdAn3XFz3xI6viYQwuRM03Fh7yXtcG4nx.dzGemcTP7\r\n 4dSP3xegGFtBO9QZni498Kcr6Mposx21DxJHZ2n6ZJ8EvGYC1xF7J_fzc9nLuGMJsgLw9zTqcVNd\r\n zW3iju1t9wB1csE9ASQVTKkHh4nsBzqm1IFUI4QlMbTX7pf7NIDoOJzbB2QRegrNuUXoIjdqmkd0\r\n ZL6Dn5DAHnrxT_NGOmD3HV0xugG56OVn2nXqPnnZzBy_7y8WOJxGYlZkWzNoO3DKTnYsw9vnCGNK\r\n 0C5x1L0dOpuzCYyTk6xoCSsf_oQXym_IuWccTMEuQHKfG2hdoxe32Iekv_aPDQjctpHHWVCDVTI0\r\n bGDXPToQfsDdMg.4WXBGUKm.kL.DkWkVAM3TiiOiqux.LspOxSAdEHAOwTAiNlTFaJZ0VZvsjYno\r\n y9XI7Fsa.dBI.Cn0fI22bz9GgcXZQ7OHKSqoo9ocIpDjl.sW3jkZUHY1QDSSCS1.4jzdu1aG1PDp\r\n mDqyCiOL9lJtJ9lXkux0vJu3Mqf2QZRq80vSDSMvhGO.UcupFoaLYh1HNzvbacoLTPDng5Lt6d7m\r\n Yjy27xTC3oPyfrYkcOlCjPm8u2q.L1a8yTVhaGY1DAF_XYiYmpiuTKWZjg0HbqwsOWrdwkgmtFUQ\r\n 97c3stRkuKDRbnyTjkpUZZOtQCUJJvJpSX9WNvk4Qf91cNnPMX_YxdReAxvNr1xkXIPAXEc5J.Rk\r\n 0P_IN_.TvPFvb6jTIwpT4TKFpPB2nEZJ8N4.REX7x3xwjofYHdWgBXfB5nqocw1KQcwolHkvN16v\r\n 2GsYFLP0HtOsdpf2jKmL345pef2GRddUdxfCENaEv0vbx_TVC1N7zZsHDWrl2ks7n2hGOP_LTPZT\r\n rRPOaUaOhmjoM6AJtgfL4N3MlIvgWcz02VNqj_G9GM8Pw.3b97vSNIAQxfNgaoJKNbyVp2ugBWe5\r\n GCbQyX9AB.nyWh6hX4ADzlJ8EkrQZRUwQXSTONckaYfeKoR6RPdczGIpaKMMghUVUWeEzL9ZUtUf\r\n n8I1VI33t4Lx0aU0Lg2b0k3AvuEMf01hsljU6VhGRwbuw7.HTW1ibJcdhhNymznfnPhVvK0Yos4J\r\n I2lguRWaEfRkm76DhoTiGNZkhIBY-\r\nX-Sonic-MF: <johnsmith@yahoo.com>\r\nX-Sonic-ID: 221dd87c-6ccb-4e96-8074-d332622b8b87\r\nReceived: from sonic.gate.mail.ne1.yahoo.com by sonic313.consmr.mail.ne1.yahoo.com with HTTP; Fri, 10 Nov 2023 09:11:08 +0000\r\nReceived: by hermes--production-ne1-56df75844-sgvl5 (Yahoo Inc. Hermes SMTP Server) with ESMTPA ID 24c441220d3992949f129e5823a987f8;\r\n          Fri, 10 Nov 2023 09:11:07 +0000 (UTC)\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\nFrom: <johnsmith@yahoo.com>\r\nMime-Version: 1.0 (1.0)\r\nSubject: Re: subject\r\nDate: Fri, 10 Nov 2023 01:10:56 -0800\r\nMessage-Id: <230A01FA-D0C6-4831-A454-FE5615AAA24A@yahoo.com>\r\nReferences: <65457bb0435d314ea86090d1@mailersend.net>\r\nIn-Reply-To: <65457bb0435d314ea86090d1@mailersend.net>\r\nTo: ben@domain.com\r\nX-Mailer: iPhone Mail (20G81)\r\nContent-Length: 84\r\n\r\nHello, world!\r\n\r\nOn Nov 3, 2023, at 16:01, ben@domain.com wrote:\r\n>=20\r\n> =EF=BB=BFTest\r\n\r\n"
		inReplyTo = "<65457bb0435d314ea86090d1@mailersend.net>"
		text      = "Hello, world!\r\n\r\nOn Nov 3, 2023, at 16:01, ben@domain.com wrote:\r\n> \r\n> \ufeffTest\r\n\r\n"
	)
	cfg := config.Config{
		Domain: "domain.com",
	}
	record := &kgo.Record{
		Value: []byte(emailData),
	}

	// emailService.ThreadSearch expectation
	thread := model.EmailThread{
		Id:           primitive.NewObjectID(),
		Participants: []usermodel.User{sender, rcpt},
		Emails: []model.Email{
			{
				MessageId: inReplyTo,
			},
		},
	}
	emailSvc.On("ThreadSearch", mock.Anything, mock.MatchedBy(func(st model.ThreadSearchTerms) bool {
		return st.EmailMessageId == inReplyTo
	})).
		Return(thread, nil)

	// mailer.Send expectation; case response.StatusCode == 202
	header := make(http.Header)
	header.Add("X-Message-Id", mailerMsgId)
	tMailer.On("Send", mock.Anything, mock.MatchedBy(func(outbound enmime.Envelope) bool {
		return outbound.GetHeader("From") == fmt.Sprintf("%s <%s@%s>",
			sender.Name(), "mailer", cfg.Domain) &&
			outbound.GetHeader("To") == fmt.Sprintf("%s <%s>",
				rcpt.Name(), rcpt.Email) &&
			outbound.GetHeader("Subject") == "Re: subject" &&
			outbound.Text == text
	})).Return(&http.Response{
		StatusCode: 202,
		Header:     header,
	}, nil)

	// mailer.GetEmail expectation
	email := model.Email{
		MessageId: fmt.Sprintf("<%s@mailersend.net>", mailerMsgId),
	}
	tMailer.On("GetEmail", mock.Anything, mailerMsgId).Return(email, nil)

	// emailService.AddEmail expectation
	emailSvc.On("AddEmail", mock.Anything, thread.Id, email).Return(nil)
	forwardEmail(context.TODO(), cfg, emailSvc, tMailer, record)

	emailSvc.AssertExpectations(t)
	tMailer.AssertExpectations(t)
}

// mocks
type emailService struct {
	mock.Mock
}

func (s *emailService) ThreadSearch(
	ctx context.Context,
	st model.ThreadSearchTerms,
) (model.EmailThread, httputil.HttpError) {
	args := s.Called(ctx, st)
	err := args.Get(1)
	if err != nil {
		return args.Get(0).(model.EmailThread), err.(httputil.HttpError)
	}
	return args.Get(0).(model.EmailThread), nil
}

func (s *emailService) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email model.Email,
) httputil.HttpError {
	args := s.Called(ctx, threadId, email)
	err := args.Get(0)
	if err != nil {
		return err.(httputil.HttpError)
	}
	return nil
}

type testMailer struct {
	mock.Mock
}

func (m *testMailer) GetEmail(
	ctx context.Context,
	id string,
) (model.Email, httputil.HttpError) {
	args := m.Called(ctx, id)
	eml := args.Get(0).(model.Email)
	err := args.Get(1)
	if err != nil {
		return eml, err.(httputil.HttpError)
	}
	return eml, nil
}

func (m *testMailer) Send(ctx context.Context, env enmime.Envelope) (*http.Response, error) {
	args := m.Called(ctx, env)
	return args.Get(0).(*http.Response), args.Error(1)
}
