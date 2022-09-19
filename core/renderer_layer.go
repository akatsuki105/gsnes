package core

import (
	"fmt"

	"github.com/pokemium/iro"
)

/*
16x16のbgタイルで左上のBGタイルをNとしたときにその他のタイルIDがどうなるか

[0, 1] → [N,    N+1 ]
[2, 3]   [N+16, N+17]
*/
var ofs16 = [4]uint16{0, 1, 16, 17}

type layer interface {
	drawScanline(buf []iro.RGB555, y uint16, start, end int)
	enable() bool
	isMain() bool
	isSub() bool
}

type backdrop struct {
	color *iro.RGB555
}

func newBackdrop(color *iro.RGB555) *backdrop {
	return &backdrop{
		color: color,
	}
}

func (b *backdrop) enable() bool {
	return true
}

func (b *backdrop) isMain() bool {
	return true
}

func (b *backdrop) isSub() bool {
	return true
}

func (b *backdrop) drawScanline(buf []iro.RGB555, y uint16, start, end int) {
	c := *b.color
	for i := start; i < end; i++ {
		buf[i] = c
	}
}

type _bg struct {
	r             *renderer
	color         uint // 4(2bpp) or 16(4bpp) or 256(8bpp)
	index         int  // 1..4 (BG1, BG2, BG3, BG4)
	mainsc, subsc bool
	size          [2]uint16 // Screen size(BGnSC), 0: X(32 or 64), 1: Y(32 or 64)
	tileDataAddr  uint      // BG12NBA, BG34NBA
	tilemapAddr   uint      // BGnSC
	tilesize      uint16    // BGMODE.4-7, 8(8x8) or 16(16x16)
	palOfs        int
	sc            [2]scroll // 0: X, 1: Y
}

type scroll struct {
	val  uint16 // 実際のスクロール値(reg&0x3FF)
	reg  uint16 // レジスタの生の値
	prev uint8
}

type bg struct {
	*_bg
	prio bool // BGMap.13
}

func newBg(r *renderer, index int) *bg {
	return &bg{
		_bg: &_bg{
			r:        r,
			index:    index,
			tilesize: 8,
		},
	}
}

func (b *bg) enable() bool {
	return b.color != DISABLED
}

func (b *bg) isMain() bool {
	return b.mainsc
}

func (b *bg) isSub() bool {
	return b.subsc
}

func (b *bg) drawScanline(buf []iro.RGB555, y uint16, start, end int) {
	bgmap := b.r.vram.buf[b.tilemapAddr/2:]
	tiledata := b.r.vram.buf[b.tileDataAddr/2:]
	scx, scy := b.sc[0].val, b.sc[1].val

	y = (y + scy) % (b.size[1] * 8)
	row := y / b.tilesize // 上から何タイル目？

	for x := start; x < end; x++ {
		xx := (uint16(x) + scx) % (b.size[0] * 8)

		/*
			16x16のときに対象がどの象限にいるか
			[0 1]
			[2 3]
		*/
		quadrant := 0
		if x%int(b.tilesize) >= 8 {
			quadrant += 1
		}
		if y%b.tilesize >= 8 {
			quadrant += 2
		}

		// (タイルサイズ関係なく)8pxずつ描画
		if xx&0b111 == 0 {
			col := uint16(xx / b.tilesize) // 左から何タイル目？

			idx := (row&0x1F)*32 + (col & 0x1F)
			if col >= 32 {
				idx += 0x800 / 2
			}
			if row >= 32 {
				idx += 0x800 * (b.size[0] / 32) / 2
			}
			entry := bgmap[idx]

			tileID := (entry & 1023) + ofs16[quadrant]
			palID := int((entry >> 10) & 0b111)
			hflip, vflip := bit(entry, 14), bit(entry, 15)

			if b.prio != bit(entry, 13) {
				continue
			}

			switch b.color {
			case COLOR_2BPP:
				ofs := b.palOfs + (4 * palID)
				pal := b.r.pal.buf[ofs : ofs+4]

				tile := tiledata[8*tileID : (8*tileID)+8] // 2bpp: 1タイル = 16バイト
				flipedY := flip(8, vflip, int(y&0b111))
				rowdata := tile[flipedY]
				planes := [2]uint8{uint8(rowdata), uint8(rowdata >> 8)}

				// 1pxずつ描画
				for i := 0; i < 8; i++ {
					colorID := uint8(0)
					for j := 0; j < 2; j++ {
						colorID += ((planes[j] >> (7 - i)) & 0b1) << j
					}
					if colorID != 0 {
						buf[x+flip(8, hflip, i)] = pal[colorID]
					}
				}

			case COLOR_4BPP:
				ofs := b.palOfs + 16*palID
				pal := b.r.pal.buf[ofs : ofs+16]

				tile := tiledata[16*tileID : (16*tileID)+16] // 4bpp: 1タイル = 32バイト
				flipedY := flip(8, vflip, int(y&0b111))
				rowdata := [2]uint16{tile[flipedY], tile[flipedY+8]}
				planes := [4]uint8{
					uint8(rowdata[0]), uint8(rowdata[0] >> 8),
					uint8(rowdata[1]), uint8(rowdata[1] >> 8),
				}

				// 1pxずつ描画
				for i := 0; i < 8; i++ {
					colorID := uint8(0)
					for j := 0; j < 4; j++ {
						colorID += ((planes[j] >> (7 - i)) & 0b1) << j
					}
					if colorID != 0 {
						buf[x+flip(8, hflip, i)] = pal[colorID]
					}
				}

			case COLOR_7BPP:
				crash("7bpp")

			case COLOR_8BPP:
				crash("8bpp")

			default:
				crash("Invalid color format: %d", b.color)
			}
		}
	}
}

func (b *bg) String() string {
	scx, scy := b.sc[0].val, b.sc[1].val
	size := fmt.Sprintf("%dx%d", b.size[0], b.size[1])
	line1 := fmt.Sprintf("Size: %s, Data: 0x%04X, Map: 0x%04X", size, b.tileDataAddr, b.tilemapAddr)
	line2 := fmt.Sprintf("  Scroll: (%d, %d), TileSize: %dpx", scx, scy, b.tilesize)
	return line1 + "\n" + line2
}

func newBgH(b *bg) *bg {
	return &bg{
		_bg:  b._bg,
		prio: true,
	}
}

type objl struct {
	r             *renderer
	index         uint8 // 0,1,2,3: OBJ0,OBJ1,OBJ2,OBJ3
	mainsc, subsc bool
}

func newObjLayer(r *renderer, index uint8) *objl {
	return &objl{
		r:     r,
		index: index,
	}
}

func (o *objl) enable() bool {
	return true
}

func (o *objl) isMain() bool {
	return o.mainsc
}

func (o *objl) isSub() bool {
	return o.subsc
}

func (o *objl) drawScanline(buf []iro.RGB555, y uint16, start, end int) {
	oam := o.r.oam

	highest := 0

	if oam.rotate {
		highest = int((oam.addr >> 2) & 0x7F) // OBJ #N
	}

	lowest := highest - 1
	if lowest == -1 {
		lowest = 127
	}
	i := lowest
	for {
		obj := &oam.objs[i]
		if obj.prio == o.index {
			top := int(obj.y)
			height := oam.size[0]
			if obj.large {
				height = oam.size[1]
			}

			if top <= int(y) && int(y) < top+height {
				o.drawObjScanline(i, buf, y, start, end)
			}
		}

		// next
		i = i - 1
		if i == -1 {
			i = 127
		}

		// all done
		if i == lowest {
			break
		}
	}
}

// Draw an obj at `row=y`
func (o *objl) drawObjScanline(i int, buf []iro.RGB555, y uint16, start, end int) {
	oam := o.r.oam
	obj := &oam.objs[i]

	pal := o.r.pal.buf[0x80:]
	pal = pal[obj.palID*16 : (obj.palID+1)*16]

	baseAddr := oam.tileDataAddr
	if obj.table2 {
		baseAddr += 0x1000 + oam.gap
	}

	tiledata := o.r.vram.buf[baseAddr:]

	width, height := oam.size[0], oam.size[0]
	if obj.large {
		width, height = oam.size[1], oam.size[1]
	}

	row := y - uint16(obj.y) // (スプライトの一番上を0行目として)上から何行目か
	if obj.vflip {
		row = uint16(height-1) - row
	}
	tilerow := row / 8                                                                   // 上から何タイル目か
	tileID := (uint16(obj.tile) & 0xFF00) | ((uint16(obj.tile) + 16*(tilerow)) & 0x00FF) // NOTE: OBJタイルの3桁目(Hex)は桁上がりしない

	for x := 0; x < width; x++ {
		if x&0b111 == 0 {
			col := uint16(x / 8)                                    // 左から何タイル目か
			tileID := (tileID & 0xFFF0) | ((tileID + col) & 0x000F) // NOTE: OBJタイルの2,3桁目(Hex)は桁上がりしない
			tile := tiledata[16*(tileID):]
			flipedY := flip(8, obj.vflip, int(y-uint16(obj.y)))
			rowdata := [2]uint16{tile[flipedY&0b111], tile[flipedY&0b111+8]}
			planes := [4]uint8{
				uint8(rowdata[0]), uint8(rowdata[0] >> 8),
				uint8(rowdata[1]), uint8(rowdata[1] >> 8),
			}

			// 1pxずつ描画
			for i := 0; i < 8; i++ {
				colorID := uint8(0)
				for j := 0; j < 4; j++ {
					colorID += ((planes[j] >> (7 - i)) & 0b1) << j
				}
				if colorID != 0 {
					buf[int(obj.x)+flip(width, obj.hflip, x+i)] = pal[colorID]
				}
			}
		}
	}
}

func flip(size int, b bool, i int) int {
	if b {
		return (size - 1) - i
	}
	return i
}
