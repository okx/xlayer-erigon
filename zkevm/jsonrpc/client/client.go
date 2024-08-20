package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/ledgerwatch/erigon/zkevm/jsonrpc/types"
	"github.com/ledgerwatch/log/v3"
)

// Client defines typed wrappers for the zkEVM RPC API.
type Client struct {
	url string
}

// NewClient creates an instance of client
func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

// HTTPError custom error type for handling HTTP responses
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("invalid status code, expected: %d, found: %d", http.StatusOK, e.StatusCode)
}

var once sync.Once
var inputCount = 0
var outCount = 0
var errorCount = 0
var socketMap map[uintptr]struct{}

func printCount() {
	log.Info(fmt.Sprintf("HTTP requests count"))
	for {
		time.Sleep(60 * time.Second)
		temp := ""
		for k := range socketMap {
			temp += fmt.Sprintf("%d,", k)
		}
		log.Info(fmt.Sprintf("HTTP requests inputCount: %d, outCount:%v, errorCount:%v, socket map:%v", inputCount, outCount, errorCount, temp))
	}
}

// JSONRPCCall executes a 2.0 JSON RPC HTTP Post Request to the provided URL with
// the provided method and parameters, which is compatible with the Ethereum
// JSON RPC Server.
func JSONRPCCall(url, method string, parameters ...interface{}) (types.Response, error) {
	once.Do(printCount)
	const jsonRPCVersion = "2.0"
	inputCount += inputCount

	params, err := json.Marshal(parameters)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}

	req := types.Request{
		JSONRPC: jsonRPCVersion,
		ID:      float64(1),
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}

	reqBodyReader := bytes.NewReader(reqBody)
	httpReq, err := http.NewRequest(http.MethodPost, url, reqBodyReader)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}

	httpReq.Header.Add("Content-type", "application/json")

	trace := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			fmt.Printf("Connection reused: %v\n", info.Reused)
			conn := info.Conn

			if tcpConn, ok := conn.(*net.TCPConn); ok {
				connFile, err := tcpConn.File()
				if err != nil {
					return
				}
				socketID := connFile.Fd()
				socketMap[socketID] = struct{}{}
				connFile.Close()
			}
		},
	}

	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}
	if httpRes.Body != nil {
		defer httpRes.Body.Close()
	}

	if httpRes.StatusCode != http.StatusOK {
		errorCount += errorCount
		return types.Response{}, &HTTPError{StatusCode: httpRes.StatusCode}
	}

	resBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}
	//defer httpRes.Body.Close()

	var res types.Response
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		errorCount += errorCount
		return types.Response{}, err
	}
	outCount += outCount
	return res, nil
}

func JSONRPCBatchCall(url string, methods []string, parameterGroups ...[]interface{}) ([]types.Response, error) {
	const jsonRPCVersion = "2.0"

	if len(methods) != len(parameterGroups) {
		return nil, fmt.Errorf("methods and parameterGroups must have the same length")
	}

	batchRequest := make([]types.Request, 0, len(methods))

	for i, method := range methods {
		params, err := json.Marshal(parameterGroups[i])
		if err != nil {
			return nil, err
		}

		req := types.Request{
			JSONRPC: jsonRPCVersion,
			ID:      float64(i + 1),
			Method:  method,
			Params:  params,
		}

		batchRequest = append(batchRequest, req)
	}

	reqBody, err := json.Marshal(batchRequest)
	if err != nil {
		return nil, err
	}

	reqBodyReader := bytes.NewReader(reqBody)
	httpReq, err := http.NewRequest(http.MethodPost, url, reqBodyReader)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Content-type", "application/json")

	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code, expected: %v, found: %v", http.StatusOK, httpRes.StatusCode)
	}

	resBody, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	var batchResponse []types.Response
	err = json.Unmarshal(resBody, &batchResponse)
	if err != nil {
		return nil, err
	}

	return batchResponse, nil
}
