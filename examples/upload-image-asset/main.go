package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/tilebox/loops-go"
)

const contentType = "image/png"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	client, err := loops.NewClient(loops.WithAPIKey("YOUR_LOOPS_API_KEY"))
	if err != nil {
		slog.Error("failed to create client", slog.Any("error", err.Error()))
		return
	}

	ctx := context.Background()
	image, err := examplePNG()
	if err != nil {
		slog.Error("failed to decode example image", slog.Any("error", err.Error()))
		return
	}

	upload, err := client.CreateUpload(ctx, &loops.CreateUploadRequest{
		ContentType:   contentType,
		ContentLength: len(image),
	})
	if err != nil {
		slog.Error("failed to create upload", slog.Any("error", err.Error()))
		return
	}
	slog.Info("created upload", slog.String("emailAssetId", upload.EmailAssetID))

	if err := putPresignedURL(ctx, upload.PresignedURL, image); err != nil {
		slog.Error("failed to upload image bytes", slog.Any("error", err.Error()))
		return
	}
	slog.Info("uploaded image bytes", slog.Int("bytes", len(image)))

	completed, err := client.CompleteUpload(ctx, upload.EmailAssetID)
	if err != nil {
		slog.Error("failed to complete upload", slog.Any("error", err.Error()))
		return
	}
	slog.Info("completed upload", slog.String("emailAssetId", completed.EmailAssetID), slog.String("url", completed.FinalURL))
}

func examplePNG() ([]byte, error) {
	// A 1×1 transparent PNG.
	return base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
}

func putPresignedURL(ctx context.Context, url string, data []byte) error {
	// The URL is returned by Loops' CreateUpload API and points to a presigned object-storage upload URL.

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(data))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %s: %s", resp.Status, string(body))
	}
	return nil
}
