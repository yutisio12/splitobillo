package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultPort    = "8090"
	defaultLang    = "eng+ind"
	maxUploadBytes = 10 << 20
)

type extractRequest struct {
	lang string
	data []byte
}

func main() {
	logger := log.New(os.Stdout, "ocr: ", log.LstdFlags|log.Lmsgprefix)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /extract", handleExtract)

	srv := &http.Server{
		Addr:              ":" + port(),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Printf("ocr service started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("server error: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Printf("ocr service stopped")
}

func port() string {
	if p := os.Getenv("OCR_PORT"); p != "" {
		return p
	}
	return defaultPort
}

func lang() string {
	if l := os.Getenv("OCR_LANG"); l != "" {
		return l
	}
	return defaultLang
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "OCR_UNAVAILABLE", "message": "tesseract binary not found"},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "status": "ok"})
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	req, err := parseRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "INVALID_REQUEST", "message": err.Error()},
		})
		return
	}

	img, _, err := image.Decode(bytes.NewReader(req.data))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "INVALID_IMAGE", "message": "unable to decode image"},
		})
		return
	}

	tmp, err := os.CreateTemp("", "ocr-*.png")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "INTERNAL_ERROR", "message": "failed to create temp file"},
		})
		return
	}
	defer os.Remove(tmp.Name())

	if err := png.Encode(tmp, toGrayscale(img)); err != nil {
		tmp.Close()
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "INTERNAL_ERROR", "message": "failed to encode preprocessed image"},
		})
		return
	}
	tmp.Close()

	out, err := exec.Command("tesseract", tmp.Name(), "stdout", "-l", req.lang).Output()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   map[string]string{"code": "OCR_FAILED", "message": "tesseract execution failed: " + err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"text":        string(out),
			"lang":        req.lang,
			"duration_ms": time.Since(start).Milliseconds(),
		},
	})
}

func parseRequest(r *http.Request) (extractRequest, error) {
	var req extractRequest
	req.lang = lang()

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			return req, err
		}
		f, _, err := r.FormFile("file")
		if err != nil {
			return req, errors.New("missing multipart file field 'file'")
		}
		defer f.Close()
		req.data, err = readLimited(f, maxUploadBytes)
		if err != nil {
			return req, err
		}
		return req, nil
	}

	data, err := readLimited(r.Body, maxUploadBytes)
	if err != nil {
		return req, err
	}
	req.data = data
	return req, nil
}

func readLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, errors.New("file too large")
	}
	return data, nil
}

func toGrayscale(img image.Image) *image.Gray {
	b := img.Bounds()
	gray := image.NewGray(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			y8 := (299*r + 587*g + 114*bl) / 1000
			gray.Set(x, y, color.Gray{Y: uint8(y8 >> 8)})
		}
	}
	return gray
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
