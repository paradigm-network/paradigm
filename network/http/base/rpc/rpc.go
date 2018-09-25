package rpc

import (
	"encoding/json"
	"net/http"
	"sync"
	"io/ioutil"
	"github.com/rs/zerolog/log"
	"github.com/paradigm-network/paradigm/core"
)

func init() {
	mainMux.m = make(map[string]func(w http.ResponseWriter, r *http.Request))
}

//an instance of the multiplexer
var mainMux ServeMux

//multiplexer that keeps track of every function to be called on specific rpc call
type ServeMux struct {
	sync.RWMutex
	m map[string]func(w http.ResponseWriter, r *http.Request)
	defaultFunction func(http.ResponseWriter, *http.Request)

	node *core.Node
}

// this is the function that should be called in order to answer an rpc call
// should be registered like "http.HandleFunc("/", httpjsonrpc.Handle)"
func Handle(w http.ResponseWriter, r *http.Request) {

	mainMux.RLock()
	defer mainMux.RUnlock()
	if r.Method == "OPTIONS" {
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("content-type", "application/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}

	//JSON RPC commands should be POSTs
	if r.Method != "POST" {
		if mainMux.defaultFunction != nil {
			log.Info().Msg("HTTP JSON RPC Handle - Method!=\"POST\"")
			mainMux.defaultFunction(w, r)
			return
		} else {
			log.Warn().Msg("HTTP JSON RPC Handle - Method!=\"POST\"")
			return
		}
	}

	//check if there is Request Body to read
	if r.Body == nil {
		if mainMux.defaultFunction != nil {
			log.Info().Msg("HTTP JSON RPC Handle - Request body is nil")
			mainMux.defaultFunction(w, r)
			return
		} else {
			log.Warn().Msg("HTTP JSON RPC Handle - Request body is nil")
			return
		}
	}

	//read the body of the request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Msg("HTTP JSON RPC Handle - ioutil.ReadAll error")
		return
	}
	request := make(map[string]interface{})
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Error().Msg("HTTP JSON RPC Handle - json.Unmarshal error")
		return
	}
	if request["method"] == nil {
		log.Error().Msg("HTTP JSON RPC Handle - method not found")
		return
	}
	method, ok := request["method"].(string)
	if !ok {
		log.Error().Msg("HTTP JSON RPC Handle - method is not string")
		return
	}
	//get the corresponding function
	_, ok = mainMux.m[method]
	if ok {
		return
	} else {
		//if the function does not exist
		log.Warn().Msg("HTTP JSON RPC Handle - No function to call for ")
	}
}

//a function to register functions to be called for specific rpc calls
func HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	mainMux.Lock()
	defer mainMux.Unlock()
	mainMux.m[pattern] = handler
}
