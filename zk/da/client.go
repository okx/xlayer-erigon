package da

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/log/v3"

	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
)

const maxAttempts = 60
const retryDelay = 1000 * time.Millisecond

func GetOffChainData(ctx context.Context, url string, hash common.Hash) ([]byte, error) {
	attemp := 0

	for attemp < maxAttempts {
		// X Layer for timeout
		http.DefaultClient.Timeout = 30 * time.Second
		transport := &http.Transport{
			DisableKeepAlives: false,
		}
		http.DefaultClient.Transport = transport
		response, err := client.JSONRPCCall(url, "sync_getOffChainData", hash)

		if httpErr, ok := err.(*client.HTTPError); ok && httpErr.StatusCode == http.StatusTooManyRequests {
			log.Error(fmt.Sprintf("GetOffChainData StatusTooManyRequests， hash:%v, attemp:%v, error:%v", hash.String(), attemp, err))
			time.Sleep(retryDelay)
			attemp += 1
			continue
		}

		if err == io.EOF {
			log.Error(fmt.Sprintf("GetOffChainData io.EOF， hash:%v, attemp:%v, error:%v", hash.String(), attemp, err))
			time.Sleep(retryDelay)
			attemp += 1
			continue
		}

		if err != nil {
			return nil, err
		}

		if response.Error != nil {
			return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
		}

		return hexutil.Decode(strings.Trim(string(response.Result), "\""))
	}

	return nil, fmt.Errorf("max attempts of data fetching reached, attempts: %v, DA url: %s", maxAttempts, url)
}
