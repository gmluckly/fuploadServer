package service

import "strconv"

func Run(port int) {
	r := setUpRouter()
	portStr := strconv.Itoa(port)
	r.Run(":" + portStr)
}
