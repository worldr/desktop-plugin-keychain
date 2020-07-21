package keychain

import (
	"log"
	"strings"

	"github.com/go-flutter-desktop/go-flutter"
	"github.com/go-flutter-desktop/go-flutter/plugin"
	"github.com/pkg/errors"
	"github.com/zalando/go-keyring"
)

const channelName = "com.worldr.desktop/keychain"

const (
	METHOD_DELETE = "deleteKey"
	METHOD_WRITE  = "writeKey"
	METHOD_READ   = "readKey"
	PARAM_KEY     = "key"
	PARAM_VALUE   = "value"
)

type KeychainPlugin struct {
	ServiceName string
}

var _ flutter.Plugin = &KeychainPlugin{} // compile-time type check
var errorFormat string = "[keychain] %v"

func NewKeychainPlugin(serviceName string) *KeychainPlugin {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	return &KeychainPlugin{
		ServiceName: serviceName,
	}
}

func (p *KeychainPlugin) InitPlugin(messenger plugin.BinaryMessenger) error {
	if p.ServiceName == "" {
		return newError("KeychainPlugin.ServiceName must be set")
	}
	channel := plugin.NewMethodChannel(messenger, channelName, plugin.StandardMethodCodec{})
	channel.HandleFunc(METHOD_DELETE, p.handleDelete)
	channel.HandleFunc(METHOD_READ, p.handleRead)
	channel.HandleFunc(METHOD_WRITE, p.handleWrite)
	return nil
}

func getKeyValue(arguments interface{}, requireValue bool) (*string, *string, error) {
	var args map[interface{}]interface{}
	var exists bool
	var k, v interface{}
	if args, exists = arguments.(map[interface{}]interface{}); !exists {
		return nil, nil, errors.New("invalid params, must be a map")
	}
	if k, exists = args[PARAM_KEY]; !exists {
		return nil, nil, errors.New("key parameter is required")
	}
	if v, exists = args[PARAM_VALUE]; requireValue && !exists {
		return nil, nil, errors.New("value parameter is required")
	}
	var ok bool
	var key, value string
	if key, ok = k.(string); !ok || len(key) == 0 {
		return nil, nil, errors.New("key parameter must be a non-empty string")
	}
	if v != nil {
		if value, ok = v.(string); !ok {
			return nil, nil, errors.New("value parameter, if present, must be a non-empty string")
		}
	}
	return &key, &value, nil
}

func (p *KeychainPlugin) handleDelete(arguments interface{}) (reply interface{}, err error) {
	k, _, err := getKeyValue(arguments, false)
	if err != nil {
		return nil, newError(err.Error())
	}
	return nil, keyring.Delete(p.keyLabel(k), *k)
}

func (p *KeychainPlugin) handleRead(arguments interface{}) (reply interface{}, err error) {
	k, _, err := getKeyValue(arguments, false)
	if err != nil {
		return nil, newError(err.Error())
	}
	v, err := keyring.Get(p.keyLabel(k), *k)
	if err != nil {
		if strings.Index(err.Error(), "not found") >= 0 {
			// not found is fine
			return nil, nil
		} else {
			// other errors must be reported
			return nil, err
		}
	} else {
		return v, nil
	}
}

func (p *KeychainPlugin) handleWrite(arguments interface{}) (reply interface{}, err error) {
	k, v, err := getKeyValue(arguments, true)
	if err != nil {
		return nil, newError(err.Error())
	}
	return nil, keyring.Set(p.keyLabel(k), *k, *v)
}

func (p *KeychainPlugin) keyLabel(key *string) string {
	return p.ServiceName + "." + *key
}

func newError(message string) error {
	log.Printf(errorFormat, message)
	return errors.New(message)
}
