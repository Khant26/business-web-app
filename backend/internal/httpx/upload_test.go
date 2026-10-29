package httpx

import "testing"

func TestImageExt(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	jpg := []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00")
	webp := []byte("RIFF\x24\x00\x00\x00WEBPVP8 ")
	if ImageExt(png) != ".png" || ImageExt(jpg) != ".jpg" || ImageExt(webp) != ".webp" {
		t.Fatal("expected image types to be detected")
	}
	if ImageExt([]byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>")) != "" {
		t.Fatal("svg must be rejected")
	}
	if ImageExt([]byte("GIF89a")) != "" {
		t.Fatal("gif must be rejected")
	}
}

func TestParsePageDefaults(t *testing.T) {
	p := Page{Page: 3, Size: 20}
	if p.Offset() != 40 {
		t.Fatal(p.Offset())
	}
}
