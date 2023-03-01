package schema

import (
	"errors"
	"io"
	"time"
	"unsafe"

	"github.com/fufuok/bytespool"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()

	errUnmarshal = errors.New("failed to unmarshal data")
)

type DataItem struct {
	APIName string
	IP      string
	Body    []byte
	Flag    int32
	Mark    int32
}

func (d *DataItem) Size() (s uint64) {
	{
		l := uint64(len(d.APIName))

		{
			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.IP))

		{
			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.Body))

		{
			t := l
			for t >= 0x80 {
				t >>= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 8
	return
}

func (d *DataItem) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = bytespool.New64(size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.APIName))

		{
			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.APIName)
		i += l
	}
	{
		l := uint64(len(d.IP))

		{
			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.IP)
		i += l
	}
	{
		l := uint64(len(d.Body))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.Body)
		i += l
	}
	{

		buf[i+0+0] = byte(d.Flag >> 0)

		buf[i+1+0] = byte(d.Flag >> 8)

		buf[i+2+0] = byte(d.Flag >> 16)

		buf[i+3+0] = byte(d.Flag >> 24)

	}
	{

		buf[i+0+4] = byte(d.Mark >> 0)

		buf[i+1+4] = byte(d.Mark >> 8)

		buf[i+2+4] = byte(d.Mark >> 16)

		buf[i+3+4] = byte(d.Mark >> 24)

	}
	return buf[:i+8], nil
}

func (d *DataItem) Unmarshal(buf []byte) (i uint64, err error) {
	defer func() {
		// 清除 Mark 数据
		d.MarkReset()
		if r := recover(); r != nil {
			err = errUnmarshal
		}
	}()

	i = uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.APIName = string(buf[i+0 : i+0+l])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.IP = string(buf[i+0 : i+0+l])
		i += l
	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		if uint64(cap(d.Body)) >= l {
			d.Body = d.Body[:l]
		} else {
			d.Body = bytespool.New64(l)
		}
		copy(d.Body, buf[i+0:])
		i += l
	}
	{

		d.Flag = 0 | (int32(buf[i+0+0]) << 0) | (int32(buf[i+1+0]) << 8) | (int32(buf[i+2+0]) << 16) | (int32(buf[i+3+0]) << 24)

	}
	{

		d.Mark = 0 | (int32(buf[i+0+4]) << 0) | (int32(buf[i+1+4]) << 8) | (int32(buf[i+2+4]) << 16) | (int32(buf[i+3+4]) << 24)

	}
	return i + 8, nil
}
