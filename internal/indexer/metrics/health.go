// internal/indexer/metrics/health.go
package metrics

import (
    "net/http"
    "sync/atomic"
)

var (
    isReady int32
    isLive  int32
)

func SetReady(ready bool) {
    val := int32(0)
    if ready {
        val = 1
    }
    atomic.StoreInt32(&isReady, val)
}

func SetLive(live bool) {
    val := int32(0)
    if live {
        val = 1
    }
    atomic.StoreInt32(&isLive, val)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
    if atomic.LoadInt32(&isLive) == 1 {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"alive"}`))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(`{"status":"dead"}`))
    }
}

func ReadyHandler(w http.ResponseWriter, r *http.Request) {
    if atomic.LoadInt32(&isReady) == 1 {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ready"}`))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte(`{"status":"not ready"}`))
    }
}