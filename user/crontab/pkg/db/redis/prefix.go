package redis

import "strings"

const SERVICEPREFIX = "usere_crontab_"

func GetKey(key string, parts ...string) string {
	key = SERVICEPREFIX + key
	if len(parts) == 0 {
		return key
	}
	key += "_" + strings.Join(parts, "_")
	return key
}
func GetKeyByPrefix(prefix, key string, parts ...string) string {
	key = prefix + key
	if len(parts) == 0 {
		return key
	}
	key += "_" + strings.Join(parts, "_")
	return key
}
