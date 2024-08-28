package da

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gateway-fm/cdk-erigon-lib/common"
	"github.com/ledgerwatch/erigon/common/hexutil"
	"github.com/ledgerwatch/log/v3"

	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/client"
)

const retryDelay = 1000 * time.Millisecond

func GetOffChainData(ctx context.Context, url string, hash common.Hash) ([]byte, error) {
	for {
		http.DefaultClient.Timeout = 30 * time.Second
		transport := &http.Transport{
			DisableKeepAlives: false,
		}
		http.DefaultClient.Transport = transport
		response, err := client.JSONRPCCall(url, "sync_getOffChainData", hash)

		if httpErr, ok := err.(*client.HTTPError); ok && httpErr.StatusCode == http.StatusTooManyRequests {
			log.Error(fmt.Sprintf("GetOffChainData StatusTooManyRequests， hash:%v, error:%v", hash.String(), err))
			time.Sleep(retryDelay)
			continue
		}

		if err != nil {
			log.Error(fmt.Sprintf("GetOffChainData hash:%v, error:%v", hash.String(), err))
			time.Sleep(retryDelay)
			continue
		}

		if response.Error != nil {
			log.Error(fmt.Sprintf("GetOffChainData hash:%v, error:%v", hash.String(), response.Error))
			time.Sleep(retryDelay)
			continue
		}

		return hexutil.Decode(strings.Trim(string(response.Result), "\""))
	}
}
