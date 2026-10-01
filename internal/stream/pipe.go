package stream

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/utils"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.uber.org/zap"
)

// calculateBlockSize func determines optimal block size based on the range requested.
// Smaller ranges use smaller blocks to reduce wasted bandwidth during seeks.
func calculateBlockSize(start, end int64) int64 {
	size := end - start + 1

	switch {
	case size < 512*1024: // < 512KB
		return 64 * 1024 // 64KB blocks
	case size < 4*1024*1024: // < 4MB
		return 256 * 1024 // 256KB blocks
	case size < 32*1024*1024: // < 32MB
		return 512 * 1024 // 512KB blocks
	default:
		return 1024 * 1024 // 1MB blocks by default
	}
}

// WorkerEndpoint couples a client with its own valid file location
type WorkerEndpoint struct {
	Client   *gotgproto.Client
	Location tg.InputFileLocationClass
}

// StreamPipe reads data from Telegram with concurrent prefetching. implements `io.ReadCloser`.
type StreamPipe struct {
	ctx    context.Context
	cancel context.CancelFunc
	log    *zap.Logger

	endpoints []WorkerEndpoint

	// range stuff
	start      int64
	end        int64
	blockSize  int64
	totalBytes int64

	// prefetch pipeline
	blockQueue chan []byte

	// current read state
	currentBlock []byte
	blockOffset  int64
	bytesRead    int64

	// lifecycle
	closeOnce sync.Once
}

// NewStreamPipe creates a StreamPipe with multi-bot round-robin support.
func NewStreamPipe(
	ctx context.Context,
	messageID int,
	client *gotgproto.Client,
	location tg.InputFileLocationClass,
	start, end int64,
	log *zap.Logger,
) (io.ReadCloser, error) {

	if start > end {
		return nil, fmt.Errorf("invalid range: start (%d) > end (%d)", start, end)
	}

	ctx, cancel := context.WithCancel(ctx)

	totalBytes := end - start + 1
	blockSize := calculateBlockSize(start, end)

	bufferCount := config.ValueOf.StreamBufferCount
	if bufferCount < 64 {
		bufferCount = 64
	}

	endpoints := []WorkerEndpoint{
		{Client: client, Location: location},
	}

	// Try resolving file location for all registered worker bots
	if bot.Workers != nil && len(bot.Workers.Bots) > 0 {
		for _, w := range bot.Workers.Bots {
			if w.Client == nil || w.Client == client {
				continue
			}
			wFile, err := utils.FileFromMessage(ctx, w.Client, messageID)
			if err == nil && wFile != nil && wFile.Location != nil {
				endpoints = append(endpoints, WorkerEndpoint{
					Client:   w.Client,
					Location: wFile.Location,
				})
			}
		}
	}

	p := &StreamPipe{
		ctx:        ctx,
		cancel:     cancel,
		log:        log.Named("StreamPipe"),
		endpoints:  endpoints,
		start:      start,
		end:        end,
		blockSize:  blockSize,
		totalBytes: totalBytes,
		blockQueue: make(chan []byte, bufferCount),
	}

	p.log.Info("Stream pipe initialized", zap.Int("activeBots", len(endpoints)), zap.Int64("totalBytes", totalBytes))

	// start prefetching in background
	go p.prefetch()

	return p, nil
}

// Read implements io.Reader
func (p *StreamPipe) Read(buf []byte) (n int, err error) {
	if p.bytesRead >= p.totalBytes {
		return 0, io.EOF
	}

	// need a new block?
	if p.blockOffset >= int64(len(p.currentBlock)) {
		select {
		case block, ok := <-p.blockQueue:
			if !ok {
				if p.bytesRead >= p.totalBytes {
					return 0, io.EOF
				}
				return 0, ErrPipeDrained
			}
			p.currentBlock = block
			p.blockOffset = 0
		case <-p.ctx.Done():
			return 0, p.ctx.Err()
		}
	}

	// copy available data
	n = copy(buf, p.currentBlock[p.blockOffset:])
	p.blockOffset += int64(n)
	p.bytesRead += int64(n)

	return n, nil
}

// Close implements io.Closer.
// it cancels prefetching and releases resources.
func (p *StreamPipe) Close() error {
	p.closeOnce.Do(func() {
		p.cancel()
	})
	return nil
}

// prefetch runs in a continuous worker pipeline without lock-step batching pauses
func (p *StreamPipe) prefetch() {
	defer close(p.blockQueue)

	alignedStart := p.start - (p.start % p.blockSize)
	leftTrim := p.start - alignedStart
	rightTrim := (p.end % p.blockSize) + 1
	totalBlocks := int((p.end - alignedStart + p.blockSize) / p.blockSize)

	type blockResult struct {
		idx  int
		data []byte
		err  error
	}

	// Concurrency scales with active endpoints: 3 parallel requests per bot
	concurrency := len(p.endpoints) * 3
	if concurrency < 4 {
		concurrency = 4
	}
	if concurrency > 48 {
		concurrency = 48
	}
	if totalBlocks < concurrency {
		concurrency = totalBlocks
	}
	if concurrency < 1 {
		concurrency = 1
	}

	jobs := make(chan int, totalBlocks)
	results := make(chan blockResult, 128)

	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for blockIdx := range jobs {
				select {
				case <-p.ctx.Done():
					return
				default:
				}

				blockOffset := alignedStart + int64(blockIdx)*p.blockSize
				data, err := p.downloadBlockWithRetry(blockOffset, blockIdx)
				dataLen := int64(len(data))

				if err == nil {
					if totalBlocks == 1 {
						if dataLen < rightTrim {
							rightTrim = dataLen
						}
						if leftTrim > dataLen {
							leftTrim = dataLen
						}
						data = data[leftTrim:rightTrim]
					} else if blockIdx == 0 {
						if leftTrim > dataLen {
							leftTrim = dataLen
						}
						data = data[leftTrim:]
					} else if blockIdx == totalBlocks-1 {
						if dataLen > rightTrim {
							data = data[:rightTrim]
						}
					}
				}

				select {
				case results <- blockResult{idx: blockIdx, data: data, err: err}:
				case <-p.ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		for i := 0; i < totalBlocks; i++ {
			select {
			case jobs <- i:
				if i < concurrency {
					time.Sleep(10 * time.Millisecond)
				}
			case <-p.ctx.Done():
				break
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	pending := make(map[int][]byte)
	nextBlockToEmit := 0

	for res := range results {
		if res.err != nil {
			if p.ctx.Err() == nil {
				p.log.Error("block download permanently failed", zap.Int("idx", res.idx), zap.Error(res.err))
			}
			return
		}

		pending[res.idx] = res.data

		for {
			data, found := pending[nextBlockToEmit]
			if !found {
				break
			}
			delete(pending, nextBlockToEmit)

			select {
			case p.blockQueue <- data:
				nextBlockToEmit++
				if nextBlockToEmit == totalBlocks {
					return
				}
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// downloadBlockWithRetry fetches a block with round-robin bot selection, exponential backoff, and FLOOD_WAIT handling.
func (p *StreamPipe) downloadBlockWithRetry(offset int64, blockIdx int) ([]byte, error) {
	var lastErr error
	backoff := 80 * time.Millisecond
	const maxBackoff = 3 * time.Second

	for attempt := 0; attempt < 6; attempt++ {
		if p.ctx.Err() != nil {
			return nil, p.ctx.Err()
		}

		epIndex := (blockIdx + attempt) % len(p.endpoints)
		ep := p.endpoints[epIndex]

		ctx, cancel := context.WithTimeout(p.ctx, 15*time.Second)
		data, err := p.downloadBlock(ctx, ep, offset)
		cancel()

		if err == nil {
			return data, nil
		}

		lastErr = err

		if p.ctx.Err() != nil {
			return nil, p.ctx.Err()
		}

		// Handle Telegram FLOOD_WAIT automatically
		if rpcErr, ok := tgerr.As(err); ok && rpcErr.Type == "FLOOD_WAIT" {
			waitTime := time.Duration(rpcErr.Argument)*time.Second + 300*time.Millisecond
			if waitTime > 5*time.Second {
				waitTime = 5 * time.Second
			}
			p.log.Warn("FLOOD_WAIT received on bot endpoint, cooling down", zap.Int("botIndex", epIndex), zap.Duration("wait", waitTime))
			select {
			case <-time.After(waitTime):
				continue
			case <-p.ctx.Done():
				return nil, p.ctx.Err()
			}
		}

		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-p.ctx.Done():
			return nil, p.ctx.Err()
		}
	}

	return nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, lastErr)
}

// downloadBlock fetches a single block from Telegram using the multi-connection pool.
func (p *StreamPipe) downloadBlock(ctx context.Context, ep WorkerEndpoint, offset int64) ([]byte, error) {
	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    int(p.blockSize),
		Location: ep.Location,
	}

	res, err := ep.Client.API().UploadGetFile(ctx, req)
	if err != nil {
		return nil, err
	}

	switch result := res.(type) {
	case *tg.UploadFile:
		return result.Bytes, nil
	case *tg.UploadFileCDNRedirect:
		return nil, fmt.Errorf("CDN redirect not supported (redirect to DC %d)", result.DCID)
	default:
		return nil, fmt.Errorf("unexpected response type: %T", res)
	}
}
