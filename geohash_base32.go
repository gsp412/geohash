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
	Base32 = "0123456789bcdefghjkmnpqrstuvwxyz"
)

var (
	bits32   = []int{16, 8, 4, 2, 1}
	base32   = []byte(Base32)
	unBase32 = map[int32][]byte{
		'0': {0, 0, 0, 0, 0},
		'1': {0, 0, 0, 0, 1},
		'2': {0, 0, 0, 1, 0},
		'3': {0, 0, 0, 1, 1},
		'4': {0, 0, 1, 0, 0},
		'5': {0, 0, 1, 0, 1},
		'6': {0, 0, 1, 1, 0},
		'7': {0, 0, 1, 1, 1},
		'8': {0, 1, 0, 0, 0},
		'9': {0, 1, 0, 0, 1},
		'b': {0, 1, 0, 1, 0},
		'c': {0, 1, 0, 1, 1},
		'd': {0, 1, 1, 0, 0},
		'e': {0, 1, 1, 0, 1},
		'f': {0, 1, 1, 1, 0},
		'g': {0, 1, 1, 1, 1},
		'h': {1, 0, 0, 0, 0},
		'j': {1, 0, 0, 0, 1},
		'k': {1, 0, 0, 1, 0},
		'm': {1, 0, 0, 1, 1},
		'n': {1, 0, 1, 0, 0},
		'p': {1, 0, 1, 0, 1},
		'q': {1, 0, 1, 1, 0},
		'r': {1, 0, 1, 1, 1},
		's': {1, 1, 0, 0, 0},
		't': {1, 1, 0, 0, 1},
		'u': {1, 1, 0, 1, 0},
		'v': {1, 1, 0, 1, 1},
		'w': {1, 1, 1, 0, 0},
		'x': {1, 1, 1, 0, 1},
		'y': {1, 1, 1, 1, 0},
		'z': {1, 1, 1, 1, 1},
	}
	distanceBase32 = []float64{
		20000000,
		2500000,
		630000,
		78000,
		20000,
		2400,
		610,
		76,
	}
)

// base32
// geohash length	lat bits	lng bits	lat error	lng error	km error
//      1	            2	        3	    ±23     	±23		    ±2500
//      2	            5	        5		±2.8		±5.6		±630
//      3	            7	        8		±0.70		±0.70		±78
//      4	            10	        10		±0.087		±0.18		±20
//      5	            12	        13		±0.022		±0.022		±2.4
//      6	            15	        15		±0.0027		±0.0055		±0.61
//      7	            17	        18		±0.00068	±0.00068	±0.076
//      8	            20	        20		±0.000085	±0.00017	±0.019

// 输入值：纬度，经度，精度(geohash的长度)
// 返回geoHash, 以及该点所在的区域
func EncodeBase32(latitude, longitude float64, precision int) (string, *Box, error) {
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
				ch |= bits32[bit]
				minLng = mid
			} else {
				maxLng = mid
			}
		} else {
			if mid = (minLat + maxLat) / 2; mid < latitude {
				ch |= bits32[bit]
				minLat = mid
			} else {
				maxLat = mid
			}
		}

		isEven = !isEven
		if bit < 4 {
			bit++
		} else {
			geoHash.WriteByte(base32[ch])
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

// 输入值：base32 str
// 返回值 经度 + 维度
func DecodeBase32(base32 string) (lat, lng float64, err error) {
	var ub []byte
	for _, b := range base32 {
		_ub, ok := unBase32[b]
		if !ok {
			return 0, 0, errors.New("base32 invalid")
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
func DistanceBase32(point1, point2 string) float64 {
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

	if maxPrefix > len(distanceBase32)+1 {
		return 0
	}

	return distanceBase32[maxPrefix]
}

// 计算该点（latitude, longitude）在精度precision下的邻居 -- 周围8个区域+本身所在区域
// 返回这些区域的geohash值，总共9个
func GetNeighborsBase32(latitude, longitude float64, precision int) []string {
	geohashs := make([]string, 9)

	// 本身
	geohash, b, _ := EncodeBase32(latitude, longitude, precision)
	geohashs[0] = geohash

	// 上下左右
	geohashUp, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashDown, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2, precision)
	geohashLeft, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRight, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2, (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	// 四个角
	geohashLeftUp, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashLeftDown, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2-b.Width(), precision)
	geohashRightUp, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2+b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)
	geohashRightDown, _, _ := EncodeBase32((b.MinLat+b.MaxLat)/2-b.Height(), (b.MinLng+b.MaxLng)/2+b.Width(), precision)

	geohashs[1], geohashs[2], geohashs[3], geohashs[4] = geohashUp, geohashDown, geohashLeft, geohashRight
	geohashs[5], geohashs[6], geohashs[7], geohashs[8] = geohashLeftUp, geohashLeftDown, geohashRightUp, geohashRightDown

	return geohashs
}
