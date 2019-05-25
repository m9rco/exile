package common

import "errors"

var (
	ERROR_LOCK_ALREADY_REQUIRED = errors.New("The Lock has been occupied！")
	ERROR_NO_LOCAL_IP_FOUND     = errors.New("Nic IP not found! ")
)
