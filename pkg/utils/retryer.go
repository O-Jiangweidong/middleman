package utils

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

var globalRetryer *RetryManager

type RequestInfo struct {
    ID           string            `json:"id"`
    Method       string            `json:"method"`
    URL          string            `json:"url"`
    Body         interface{}       `json:"body"`
    Headers      map[string]string `json:"headers"`
    RetryCount   int               `json:"retry_count"`
    MaxRetries   int               `json:"max_retries"`
    LastError    string            `json:"last_error"`
    LastAttempt  time.Time         `json:"last_attempt"`
    FirstAttempt time.Time         `json:"first_attempt"`
    AttemptTimes []time.Time       `json:"attempt_times,omitempty"`
    Filepath     string            `json:"filepath"`
}

type RetryManager struct {
    maxRetries    int
    checkInterval time.Duration
    client        *http.Client
    mutex         sync.Mutex
    storage       Storage
    failedDir     string
    logger        *Logger
    isFinished    bool
}

type Storage interface {
    Save(request RequestInfo)
    Load() ([]RequestInfo, error)
    Delete(request RequestInfo) error
}

type JSONFileStorage struct {
    dir    string
    mutex  sync.Mutex
    logger *Logger
}

func (s *JSONFileStorage) Save(req RequestInfo) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    filename := fmt.Sprintf("%s.json", req.ID)
    req.Filepath = filepath.Join(s.dir, filename)
    
    data, err := json.MarshalIndent(req, "", "  ")
    if err != nil {
        s.logger.Error("Request save failed: ", err.Error())
        return
    }
    
    if err = os.WriteFile(req.Filepath, data, 0644); err != nil {
        s.logger.Error("Request save failed: \n", string(data))
        return
    }
    
    s.logger.Info(fmt.Sprintf("已保存请求到文件: %s", req.Filepath))
    return
}

func (s *JSONFileStorage) Load() ([]RequestInfo, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    var requests []RequestInfo
    files, err := os.ReadDir(s.dir)
    if err != nil {
        return nil, fmt.Errorf("read json-file-storage failed: %w", err)
    }
    
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
            continue
        }
        
        filePath := filepath.Join(s.dir, file.Name())
        
        data, err := os.ReadFile(filePath)
        if err != nil {
            s.logger.Error(fmt.Sprintf("read file failed: %s", filePath), err)
            continue
        }
        var req RequestInfo
        if err = json.Unmarshal(data, &req); err != nil {
            s.logger.Error(fmt.Sprintf("json parse file failed: %s", filePath), err)
            continue
        }
        requests = append(requests, req)
    }
    
    s.logger.Debug(fmt.Sprintf("Already load %d requests", len(requests)))
    return requests, nil
}

func (s *JSONFileStorage) Delete(req RequestInfo) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if err := os.Remove(req.Filepath); err != nil {
        if os.IsNotExist(err) {
            s.logger.Info(fmt.Sprintf("文件不存在，无需删除: %s", req.Filepath))
            return nil
        }
        return fmt.Errorf("删除文件失败: %w", err)
    }
    s.logger.Info(fmt.Sprintf("已删除请求文件: %s", req.Filepath))
    return nil
}

func NewRetryManager(maxRetries int) (*RetryManager, error) {
    failedDir := "data/archive_failed"
    if err := os.MkdirAll(failedDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create dir for request-failed-archive: %v", err)
    }
    
    logger := GetLogger()
    storage := &JSONFileStorage{
        dir:    "data/retry_storage",
        logger: logger,
    }
    
    if err := os.MkdirAll(storage.dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create dir for storage: %v", err)
    }
    
    manager := &RetryManager{
        maxRetries:    maxRetries,
        checkInterval: 1 * time.Minute,
        client:        &http.Client{Timeout: 2 * time.Minute},
        storage:       storage,
        failedDir:     failedDir,
        logger:        logger,
        isFinished:    true,
    }
    return manager, nil
}

func (rm *RetryManager) Start(ctx context.Context) {
    go rm.retryWorker(ctx)
}

func (rm *RetryManager) AddFailedRequest(method, url string, headers map[string]string, body interface{}, err error) {
    now := time.Now()
    req := RequestInfo{
        ID:           uuid.New().String(),
        Method:       method,
        URL:          url,
        Body:         body,
        Headers:      headers,
        RetryCount:   0,
        MaxRetries:   rm.maxRetries,
        LastError:    err.Error(),
        LastAttempt:  now,
        FirstAttempt: now,
        AttemptTimes: []time.Time{now},
    }
    rm.storage.Save(req)
}

func (rm *RetryManager) retryWorker(ctx context.Context) {
    ticker := time.NewTicker(rm.checkInterval)
    defer ticker.Stop()
    
    rm.logger.Debug("Start worker -> [retryer]")
    
    for {
        select {
        case <-ctx.Done():
            rm.logger.Info(" Worker [retryer] is exiting.")
            return
        case <-ticker.C:
            rm.checkAndRetryRequests()
        }
    }
}

func (rm *RetryManager) checkAndRetryRequests() {
    if !rm.isFinished {
        return
    }
    
    rm.isFinished = false
    requests, err := rm.storage.Load()
    if err != nil {
        rm.logger.Error("Load request list failed.")
        rm.isFinished = true
        return
    }
    
    for _, req := range requests {
        rm.retryRequest(&req)
        if req.RetryCount >= req.MaxRetries {
            // TODO 到达最大该如何处理
            rm.archiveFailedRequest(req)
            rm.storage.Delete(req)
            continue
        }
    }
    rm.isFinished = true
}

func (rm *RetryManager) retryRequest(req *RequestInfo) {
    rm.logger.Info(fmt.Sprintf("Retry request: %s %s (%d times)",
        req.Method, req.URL, req.RetryCount+1))
    
    var resp *http.Response
    var err error
    
    body, _ := json.Marshal(req.Body)
    httpReq, _ := http.NewRequest(req.Method, req.URL, bytes.NewBuffer(body))
    for key, value := range req.Headers {
        httpReq.Header.Set(key, value)
    }
    resp, err = rm.client.Do(httpReq)
    
    if err != nil {
        req.RetryCount++
        req.LastError = err.Error()
        req.LastAttempt = time.Now()
        req.AttemptTimes = append(req.AttemptTimes, req.LastAttempt)
        rm.logger.Info(fmt.Sprintf("Retry request failed: %s %s (%d times)",
            req.Method, req.URL, req.RetryCount))
        rm.storage.Save(*req)
        return
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        req.RetryCount++
        req.LastError = fmt.Sprintf("Read response failed: %v", err)
        req.LastAttempt = time.Now()
        req.AttemptTimes = append(req.AttemptTimes, req.LastAttempt)
        
        rm.logger.Info(fmt.Sprintf("Retry request failed: %s %s (%d times)",
            req.Method, req.URL, req.RetryCount))
        rm.storage.Save(*req)
        return
    }
    
    if (resp.StatusCode >= 200 && resp.StatusCode < 300) || resp.StatusCode == 404 {
        err = rm.storage.Delete(*req)
        if err != nil {
            rm.logger.Error("Request %s delete failed.", req.URL)
        } else {
            rm.logger.Info(fmt.Sprintf("Request retry success: %s %s", req.Method, req.URL))
        }
    } else {
        req.RetryCount++
        req.LastError = fmt.Sprintf("Request error: %d - %s",
            resp.StatusCode, string(respBody))
        req.LastAttempt = time.Now()
        req.AttemptTimes = append(req.AttemptTimes, req.LastAttempt)
        
        rm.logger.Info(fmt.Sprintf("Retry request failed: %s %s (%d times)",
            req.Method, req.URL, req.RetryCount))
        rm.storage.Save(*req)
    }
}

func (rm *RetryManager) archiveFailedRequest(req RequestInfo) {
    timestamp := time.Now().Format("20060102_150405")
    fileName := fmt.Sprintf("%s_%s_%s.json",
        timestamp, req.Method, filepath.Base(req.URL))
    
    filePath := filepath.Join(rm.failedDir, fileName)
    
    file, err := os.Create(filePath)
    if err != nil {
        rm.logger.Error(fmt.Sprintf("创建归档文件失败: %s", filePath), err)
        return
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(req); err != nil {
        rm.logger.Error(fmt.Sprintf("写入归档文件失败: %s", filePath), err)
        return
    }
    
    rm.logger.Info(fmt.Sprintf("请求已归档到文件: %s (最大重试次数: %d)",
        filePath, req.MaxRetries))
}

func GetRetryer() *RetryManager {
    if globalRetryer == nil {
        retryer, err := NewRetryManager(10)
        if err != nil {
            log.Fatalf("Start retry manager failed: %v", err)
        }
        globalRetryer = retryer
    }
    return globalRetryer
}
