package skia

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"testing"
)

// makeTestGIF 造一个 3 帧 GIF（8x8，每帧一种纯色，延时 10 个 1/100 秒 = 100ms）。
func makeTestGIF(t *testing.T) []byte {
	t.Helper()
	pal := color.Palette{
		color.RGBA{R: 255, A: 255},
		color.RGBA{G: 255, A: 255},
		color.RGBA{B: 255, A: 255},
	}
	var g gif.GIF
	for i := 0; i < 3; i++ {
		img := image.NewPaletted(image.Rect(0, 0, 8, 8), pal)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				img.SetColorIndex(x, y, uint8(i))
			}
		}
		g.Image = append(g.Image, img)
		g.Delay = append(g.Delay, 10)
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, &g); err != nil {
		t.Fatalf("生成测试 GIF 失败: %v", err)
	}
	return buf.Bytes()
}

// TestCodecDecodesAnimatedGIF：SkCodec 绑定要给出**帧数 / 尺寸 / 每帧延时**，并把
// 每一帧解成可读位图，且不同帧像素确实不同——「动画真的动了」，不是「解出 3 张一样
// 的图」（后者恰恰是此前「只取首帧」的表现）。
func TestCodecDecodesAnimatedGIF(t *testing.T) {
	codec, err := NewCodec(makeTestGIF(t))
	if err != nil {
		t.Fatalf("NewCodec 失败: %v", err)
	}
	defer codec.Release()

	if n := codec.FrameCount(); n != 3 {
		t.Fatalf("FrameCount = %d，want 3", n)
	}
	if w, h := codec.Dimensions(); w != 8 || h != 8 {
		t.Fatalf("Dimensions = %dx%d，want 8x8", w, h)
	}
	if d := codec.FrameDurationMS(0); d != 100 {
		t.Errorf("第 0 帧时长 = %d ms，want 100", d)
	}

	frames, err := codec.DecodeFrames()
	if err != nil {
		t.Fatalf("DecodeFrames 失败: %v", err)
	}
	if len(frames) != 3 {
		t.Fatalf("解出 %d 帧，want 3", len(frames))
	}
	seen := make([][]byte, 0, 3)
	for i, f := range frames {
		if f.Image == nil {
			t.Fatalf("第 %d 帧为空", i)
		}
		if f.DurationMS != 100 {
			t.Errorf("第 %d 帧时长 = %d ms，want 100", i, f.DurationMS)
		}
		px, err := f.Image.ReadPixels()
		if err != nil {
			t.Fatalf("第 %d 帧读像素失败: %v", i, err)
		}
		if len(px) == 0 {
			t.Fatalf("第 %d 帧像素为空", i)
		}
		for j, prev := range seen {
			if bytes.Equal(prev, px) {
				t.Errorf("第 %d 帧与第 %d 帧像素完全相同（动画没动）", i, j)
			}
		}
		seen = append(seen, px)
	}
}

// TestCodecSingleFramePNG：静态图也能建 codec（帧数 = 1）——引擎侧据此用同一套
// 代码处理「GIF/WebP 动画」与「普通 PNG/JPEG」，不必分两条路。
func TestCodecSingleFramePNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 10, G: 200, B: 30, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("编码 PNG 失败: %v", err)
	}
	codec, err := NewCodec(buf.Bytes())
	if err != nil {
		t.Fatalf("NewCodec(PNG) 失败: %v", err)
	}
	defer codec.Release()
	// ★ Skia 对静态容器不提供帧计数（返回 0），所以这里允许 0 或 1；真正要保证的是
	// DecodeFrames() 总能给出至少一帧（上层用同一套路径处理动画与静态图）。
	if n := codec.FrameCount(); n != 0 && n != 1 {
		t.Errorf("静态 PNG 的 FrameCount = %d，want 0（Skia 语义）或 1", n)
	}
	if _, err := codec.DecodeFrame(0); err != nil {
		t.Errorf("DecodeFrame(0) 失败: %v", err)
	}
	frames, err := codec.DecodeFrames()
	if err != nil {
		t.Fatalf("静态 PNG 的 DecodeFrames 失败: %v", err)
	}
	if len(frames) != 1 {
		t.Errorf("静态 PNG 应解出 1 帧（FrameCount=0 当单帧），got %d", len(frames))
	}
	if _, err := NewCodec([]byte("not an image")); err == nil {
		t.Error("非法数据应报错")
	}
}
