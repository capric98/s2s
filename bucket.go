package s2s

import (
	"context"
	"errors"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

type Buk struct {
	client    *storage.Client
	name      string
	projectID string

	ctx     context.Context
	cancel  func()
	timeout time.Duration
}

func OpenBuk(bukname, projectID string) (b *Buk, e error) {
	// Init client.
	b = &Buk{
		name:      bukname,
		projectID: projectID,
		timeout:   5 * time.Second,
	}
	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.client, e = storage.NewClient(b.ctx)
	if e != nil {
		return
	}

	// Check whether bucket exists or not.
	tctx, cancel := context.WithTimeout(b.ctx, b.timeout)
	defer cancel()
	it := b.client.Bucket(bukname).Objects(tctx, nil)
	if _, e = it.Next(); e == storage.ErrBucketNotExist {
		return b, b.create()
	}

	e = nil
	return
}

func (b *Buk) Close() error {
	b.cancel()
	return b.client.Close()
}

func (b *Buk) Upload(obj string, r io.Reader) (e error) {
	ctx, cancel := context.WithCancel(b.ctx)
	defer cancel()

	wc := b.client.Bucket(b.name).Object(obj).NewWriter(ctx)
	// This would make sure that the following goroutine could return.
	defer wc.Close()

	ec := make(chan error)
	go func() {
		var err error
		for {
			_, err = io.CopyN(wc, r, 64*1024)
			select {
			case <-ctx.Done():
				return
			case ec <- err:
			}
		}
	}()

	for {
		select {
		case <-time.After(b.timeout):
			e = errors.New("Bucket object upload timeout.")
			return
		case e = <-ec:
			switch e {
			case nil:
			case io.EOF:
				e = nil
				return
			default:
				return
			}
		}
	}
}

func (b *Buk) Delete(obj string) error {
	tctx, cancel := context.WithTimeout(b.ctx, b.timeout)
	defer cancel()
	return b.client.Bucket(b.name).Object(obj).Delete(tctx)
}

func (b *Buk) create() (e error) {
	bucket := b.client.Bucket(b.name)
	tctx, cancel := context.WithTimeout(b.ctx, b.timeout)
	defer cancel()
	e = bucket.Create(tctx, b.projectID, nil)
	return
}
