package utils

import (
	"gopkg.in/ini.v1"
)

type IniParser struct {
	confReader *ini.File // config reader
}

type IniParserError struct {
	err string
}

func (e *IniParserError) Error() string { return e.err }

func (current *IniParser) Load(configFileName string) error {
	conf, err := ini.Load(configFileName)
	if err != nil {
		current.confReader = nil
		return err
	}
	current.confReader = conf
	return nil
}

func (current *IniParser) GetString(section string, key string) string {
	if current.confReader == nil {
		return ""
	}

	s := current.confReader.Section(section)
	if s == nil {
		return ""
	}

	return s.Key(key).String()
}

func (current *IniParser) GetInt32(section string, key string) int32 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Int()

	return int32(value)
}

func (current *IniParser) GetUint32(section string, key string) uint32 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Uint()

	return uint32(value)
}

func (current *IniParser) GetInt64(section string, key string) int64 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Int64()
	return value
}

func (current *IniParser) GetUint64(section string, key string) uint64 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Uint64()
	return value
}

func (current *IniParser) GetFloat32(section string, key string) float32 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Float64()
	return float32(value)
}

func (current *IniParser) GetFloat64(section string, key string) float64 {
	if current.confReader == nil {
		return 0
	}

	s := current.confReader.Section(section)
	if s == nil {
		return 0
	}

	value, _ := s.Key(key).Float64()
	return value
}
