/**
 * @File: distance
 * @Author: Shuangpeng.Guo
 * @Date: 2021/6/29 5:00 下午
 */
package geohash

import "math"

const EarthRadius = 6371393 // m 地球半径 平均值，千米

func Distance(lng1, lat1, lng2, lat2 float64) float64 {
	// 经典计算方式
	if (math.Abs(lat1) > 90) || (math.Abs(lat2) > 90) {
		return -1
	}
	if (math.Abs(lng1) > 180) || (math.Abs(lng2) > 180) {
		return -1
	}

	// 经度换算弧度
	radLat1 := rad(lat1)
	radLat2 := rad(lat2)

	// 经度弧度差值
	vLat := math.Abs(radLat1 - radLat2)
	// 维度弧度差值
	vLng := rad(lng1) - rad(lng2)

	a := math.Sin(vLat/2)*math.Sin(vLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*math.Sin(vLng/2)*math.Sin(vLng/2)

	distance := 2 * EarthRadius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return distance
}

func rad(d float64) float64 {
	return d * math.Pi / 180.0
}
