/**
 * @File: geohash_go
 * @Author: Shuangpeng.Guo
 * @Date: 2021/6/9 3:13 下午
 */
package geohash

import (
	"bytes"
	"errors"
	"fmt"
	"math"
)

const (
	Base4                = "0123"
	MaxLatitude  float64 = 90
	MinLatitude  float64 = -90
	MaxLongitude float64 = 180
	MinLongitude float64 = -180
)

var (
	bits4   = []int{2, 1}
	base4   = []byte(Base4)
	unBase4 = map[int32][]byte{
		'0': {0, 0},
		'1': {0, 1},
		'2': {1, 0},
		'3': {1, 1},
	}
	distanceBase4 = []float64{
		20000000,
		10000000,
		5000000,
		2500000,
		1250000,
		630000,
		315000,
		157000,
		78000,
		39000,
		20000,
		9728,
		4864,
		2432,
		1216,
		608,
		304,
		152,
		76,
		38,
		19,
		9.5,
		4.75,
		2.37,
		1.18,
	}
)

// 改进base32缩小精度
// geoHash有效位2n(0<n<32)   lat bits    lng bits    lat error   lng error   km error
// 2(n=1)                   1           1           ±45         ±90         ±10000
// 4(n=2)                   2           2           ±22.5       ±45         ±5000
// 6(n=3)                   3           3           ±11.2       ±22.5       ±2500
// 8(n=4)                   4           4           ±5.6        ±11.2       ±1250
// 10(n=5)                  5           5           ±2.8        ±5.6        ±630
// 12(n=6)                  6           6           ±1.4        ±2.8        ±315
// 14(n=7)                  7           7           ±0.7        ±1.4        ±157
// 16(n=8)                  8           8           ±0.352      ±0.7        ±78
// 18(n=9)                  9           9           ±0.176      ±0.352      ±39
// 20(n=10)                 10          10          ±0.088      ±0.176      ±20
// 22(n=11)                 11          11          ±0.044      ±0.088      ±9.728
// 24(n=12)                 12          12          ±0.022      ±0.044      ±4.864
// 26(n=13)                 13          13          ±0.011      ±0.022      ±2.432
// 28(n=14)                 14          14          ±0.0055     ±0.011      ±1.216
// 30(n=15)                 15          15          ±0.00275    ±0.0055     ±0.608
// 32(n=16)                 16          16          ±0.00137    ±0.00275    ±0.304
// 34(n=17)                 17          17          ±0.00069    ±0.00137    ±0.152
// 36(n=18)                 18	        18          ±0.00034    ±0.00069    ±0.076
// 38(n=19)                 19          19          ±0.00017    ±0.00034    ±0.038
// 40(n=20)                 20          20	        ±0.000085	±0.00017	±0.019
// 42(n=21)                 21          21          ±0.0000425  ±0.000085   ±0.0095
// 44(n=22)                 22          22          ±0.00002125 ±0.0000425  ±0.00475
// 46(n=23)                 23          23          ±0.00001062 ±0.00002125 ±0.00237
// 48(n=24)                 24          24          ±0.0000053  ±0.00001062 ±0.00118

func (b *Box) Width() float64 {
	return b.MaxLng - b.MinLng
}

func (b *Box) Height() float64 {
	return b.MaxLat - b.MinLat
}

// 只留作记录 并不能很好的保证准确性
// 输入值：纬度，经度，精度(geohash的长度)
// 返回geoHash, 以及该点所在的区域
func EncodeBase4(latitude, longitude float64, precision int) (string, *Box, error) {
	var geoHash bytes.Buffer
	var minLat, maxLat float64 = MinLatitude, MaxLatitude
	var minLng, maxLng float64 = MinLongitude, MaxLongitude
	var mid float64 = 0
	if math.Abs(latitude) > MaxLatitude {
		return "", nil, errors.New("latitude overrun")
	}
	if math.Abs(longitude) > MaxLongitude {
		return "", nil, errors.New("longitude overrun")
	}

	bit, ch, length, isEven := 0, 0, 0, true
	for length < precision {
		if isEven {
			if mid = (minLng + maxLng) / 2; mid < longitude {
				ch |= bits4[bit]
				minLng = mid
			} else {
				maxLng = mid
			}
		} else {
			if mid = (minLat + maxLat) / 2; mid < latitude {
				ch |= bits4[bit]
				minLat = mid
			} else {
				maxLat = mid
			}
		}

		isEven = !isEven
		if bit < 1 {
			bit++
		} else {
			geoHash.WriteByte(base4[ch])
			length, bit, ch = length+1, 0, 0
		}
	}

	b := &Box{
		MinLat: minLat,
		MaxLat: maxLat,
		MinLng: minLng,
		MaxLng: maxLng,
	}
	fmt.Println()

	return geoHash.String(), b, nil
}

// 输入值：base4 str
// 返回值 经度 + 维度
func DecodeBase4(base4 string) (lat, lng float64, err error) {
	var ub []byte
	for _, b := range base4 {
		_ub, ok := unBase4[b]
		if !ok {
			return 0, 0, errors.New("base4 invalid")
		}
		ub = append(ub, _ub...)
	}
	isEven := true
	var minLat, maxLat float64 = MinLatitude, MaxLatitude
	var minLng, maxLng float64 = MinLongitude, MaxLongitude

	lat = (MinLatitude + MaxLatitude) / 2
	lng = (MinLongitude + MaxLongitude) / 2
	for _, v := range ub {
		if isEven {
			if v == 0 {
				maxLng = lng
				lng = (minLng + lng) / 2
			} else {
				minLng = lng
				lng = (lng + maxLng) / 2
			}
		} else {
			if v == 0 {
				maxLat = lat
				lat = (minLat + lat) / 2
			} else {
				minLat = lat
				lat = (lat + maxLat) / 2
			}
		}

		isEven = !isEven
	}
	return lat, lng, nil
}

// 输入值：point1，point2
// 返回值：两点间估计距离
func DistanceBase4(point1, point2 string) float64 {
	l := len(point1)
	if len(point2) < l {
		l = len(point2)
	}
	maxPrefix := 0
	for i := 0; i < l; i++ {
		if point1[i] != point2[i] {
			maxPrefix = i
			break
		}
	}

	if maxPrefix > len(distanceBase4)+1 {
		return 0
	}

	return distanceBase4[maxPrefix]
}

// 计算该点（latitude, longitude）在精度precision下的邻居 -- 周围8个区域+本身所在区域
// 返回这些区域的geohash值，总共9个
func GetNeighborsBase4(latitude, longitude float64, precision int) []string {
	geohashs := make([]string, 9)

	// 本身
	geohash, b, _ := EncodeBase4(latitude, longitude, precision)
	geohashs[0] = geohash

	// 上下左右
	geohashUp, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashDown, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashLeft, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRight, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	// 四个角
	geohashLeftUp, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashLeftDown, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRightUp, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)
	geohashRightDown, _, _ := EncodeBase4((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	geohashs[1], geohashs[2], geohashs[3], geohashs[4] = geohashUp, geohashDown, geohashLeft, geohashRight
	geohashs[5], geohashs[6], geohashs[7], geohashs[8] = geohashLeftUp, geohashLeftDown, geohashRightUp, geohashRightDown

	return geohashs
}
