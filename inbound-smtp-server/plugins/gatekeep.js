// gatekeep
const fetch = require('node-fetch');

exports.register = function () {
    this.cfg = this.config.get("gatekeep.ini").main;
    this.loginfo("cfg: " + JSON.stringify(this.cfg));
}

exports.hook_data_post = async function (next, conn) {
    // only replies to server sent emails qualify as valid inbound emails.
    // gatekeep verifies the "In-Reply-To" header against the emailThreads collection.
    const inReplyTo = conn?.transaction?.header?.get("In-Reply-To");
    if (inReplyTo) {
        const body = JSON.stringify({emailMessageId: inReplyTo.trim()});
        // this.loginfo("body: " + body);
        const reqOpts = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body,
        };
        const resp = await fetchPlus(this.cfg.uri + "/email/thread/search", reqOpts, 3);
        if (resp && resp.status == 200) {
            return next();
        }
    }
    
    next(DENY);
}

async function fetchPlus(url, options = {}, retries) {
  let resp;
  try {
    resp = await fetch(url, options);
    if (resp.ok) {
      return new Promise((resolve, _) => resolve(resp));
    }
  
    if (retries > 0 && resp.status > 499) {
      await delay(5000);
      return fetchPlus(url, options, retries - 1);
    }
  } catch (e) {
      await delay(5000);
      return fetchPlus(url, options, retries - 1);
  }
}

async function delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}