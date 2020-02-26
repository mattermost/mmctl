// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

const defaultEditor = "vi"

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration",
}

var ConfigGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get config setting",
	Long:    "Gets the value of a config setting by its name in dot notation.",
	Example: `config get SqlSettings.DriverName`,
	Args:    cobra.ExactArgs(1),
	RunE:    withClient(configGetCmdF),
}

var ConfigSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set config setting",
	Long:    "Sets the value of a config setting by its name in dot notation. Accepts multiple values for array settings",
	Example: "config set SqlSettings.DriverName mysql",
	Args:    cobra.MinimumNArgs(2),
	RunE:    withClient(configSetCmdF),
}

var ConfigEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit the config",
	Long:    "Opens the editor defined in the EDITOR environment variable to modify the server's configuration and then uploads it",
	Example: "config edit",
	Args:    cobra.NoArgs,
	RunE:    withClient(configEditCmdF),
}

var ConfigResetCmd = &cobra.Command{
	Use:     "reset",
	Short:   "Reset config setting",
	Long:    "Resets the value of a config setting by its name in dot notation or a setting section. Accepts multiple values for array settings.",
	Example: "config reset SqlSettings.DriverName LogSettings",
	Args:    cobra.MinimumNArgs(1),
	RunE:    withClient(configResetCmdF),
}

var ConfigShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Writes the server configuration to STDOUT",
	Long:    "Prints the server configuration and writes to STDOUT in JSON format.",
	Example: "config show",
	Args:    cobra.NoArgs,
	RunE:    withClient(configShowCmdF),
}

func init() {
	ConfigResetCmd.Flags().Bool("confirm", false, "Confirm you really want to reset all configuration settings to its default value")
	ConfigCmd.AddCommand(
		ConfigGetCmd,
		ConfigSetCmd,
		ConfigEditCmd,
		ConfigResetCmd,
		ConfigShowCmd,
	)
	RootCmd.AddCommand(ConfigCmd)
}

func getValue(path []string, obj interface{}) (interface{}, bool) {
	r := reflect.ValueOf(obj)
	var val reflect.Value
	if r.Kind() == reflect.Map {
		val = r.MapIndex(reflect.ValueOf(path[0]))
		if val.IsValid() {
			val = val.Elem()
		}
	} else {
		val = r.FieldByName(path[0])
	}

	if !val.IsValid() {
		return nil, false
	}

	switch {
	case len(path) == 1:
		return val.Interface(), true
	case val.Kind() == reflect.Struct:
		return getValue(path[1:], val.Interface())
	case val.Kind() == reflect.Map:
		remainingPath := strings.Join(path[1:], ".")
		mapIter := val.MapRange()
		for mapIter.Next() {
			key := mapIter.Key().String()
			if strings.HasPrefix(remainingPath, key) {
				i := strings.Count(key, ".") + 2 // number of dots + a dot on each side
				mapVal := mapIter.Value()
				// if no sub field path specified, return the object
				if len(path[i:]) == 0 {
					return mapVal.Interface(), true
				}
				data := mapVal.Interface()
				if mapVal.Kind() == reflect.Ptr {
					data = mapVal.Elem().Interface() // if value is a pointer, dereference it
				}
				// pass subpath
				return getValue(path[i:], data)
			}
		}
	}
	return nil, false
}

func setValueWithConversion(val reflect.Value, newValue interface{}) error {
	switch val.Kind() {
	case reflect.Struct:
		val.Set(reflect.ValueOf(newValue))
		return nil
	case reflect.Slice:
		if val.Type().Elem().Kind() != reflect.String {
			return errors.New("unsupported type of slice")
		}
		val.Set(reflect.ValueOf(newValue))
		return nil
	case reflect.Int:
		v, err := strconv.ParseInt(newValue.(string), 10, 64)
		if err != nil {
			return errors.New("target value is of type Int and provided value is not")
		}
		val.SetInt(v)
		return nil
	case reflect.String:
		val.SetString(newValue.(string))
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(newValue.(string))
		if err != nil {
			return errors.New("target value is of type Bool and provided value is not")
		}
		val.SetBool(v)
		return nil
	default:
		return errors.New("target value type is not supported")
	}
}

func setValue(path []string, obj reflect.Value, newValue interface{}) error {
	var val reflect.Value
	switch obj.Kind() {
	case reflect.Struct:
		val = obj.FieldByName(path[0])
	case reflect.Map:
		val = obj.MapIndex(reflect.ValueOf(path[0]))
		if val.IsValid() {
			val = val.Elem()
		}
	default:
		val = obj
	}

	if val.Kind() == reflect.Invalid {
		return errors.New("selected path object is not valid")
	}

	if len(path) == 1 {
		if val.Kind() == reflect.Ptr {
			return setValue(path, val.Elem(), newValue)
		} else if obj.Kind() == reflect.Map {
			// since we cannot set map elements directly, we clone the value, set it, and then put it back in the map
			mapKey := reflect.ValueOf(path[0])
			subVal := obj.MapIndex(mapKey)
			if subVal.IsValid() {
				tmpVal := reflect.New(subVal.Elem().Type())
				if err := setValueWithConversion(tmpVal.Elem(), newValue); err != nil {
					return err
				}
				obj.SetMapIndex(mapKey, tmpVal)
				return nil
			}
		}
		return setValueWithConversion(val, newValue)
	}

	if val.Kind() == reflect.Struct {
		return setValue(path[1:], val, newValue)
	} else if val.Kind() == reflect.Map {
		remainingPath := strings.Join(path[1:], ".")
		mapIter := val.MapRange()
		for mapIter.Next() {
			key := mapIter.Key().String()
			if strings.HasPrefix(remainingPath, key) {
				mapVal := mapIter.Value()

				if mapVal.Kind() == reflect.Ptr {
					mapVal = mapVal.Elem() // if value is a pointer, dereference it
				}
				i := len(strings.Split(key, ".")) + 1

				if i > len(path)-1 { // leaf element
					i = 1
					mapVal = val
				}
				// pass subpath
				return setValue(path[i:], mapVal, newValue)
			}
		}
	}
	return errors.New("path object type is not supported")
}

func setConfigValue(path []string, config *model.Config, newValue []string) error {
	if len(newValue) > 1 {
		return setValue(path, reflect.ValueOf(config).Elem(), newValue)
	}
	return setValue(path, reflect.ValueOf(config).Elem(), newValue[0])
}

func resetConfigValue(path []string, config *model.Config, newValue interface{}) error {
	nv := reflect.ValueOf(newValue)
	if nv.Kind() == reflect.Ptr {
		switch nv.Elem().Kind() {
		case reflect.Int:
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.Itoa(*newValue.(*int)))
		case reflect.Bool:
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.FormatBool(*newValue.(*bool)))
		default:
			return setValue(path, reflect.ValueOf(config).Elem(), *newValue.(*string))
		}
	} else {
		return setValue(path, reflect.ValueOf(config).Elem(), newValue)
	}
}

func configGetCmdF(c client.Client, _ *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FormatJSON)

	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path := strings.Split(args[0], ".")
	val, ok := getValue(path, *config)
	if !ok {
		return errors.New("invalid key")
	}

	printer.Print(val)
	return nil
}

func configSetCmdF(c client.Client, _ *cobra.Command, args []string) error {
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path := parseConfigPath(args[0])
	if err := setConfigValue(path, config, args[1:]); err != nil {
		return err
	}
	newConfig, res := c.UpdateConfig(config)
	if res.Error != nil {
		return res.Error
	}

	printer.PrintT("Value changed successfully", newConfig)
	return nil
}

func configEditCmdF(c client.Client, _ *cobra.Command, _ []string) error {
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile(os.TempDir(), "mmctl-*.json")
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()
	if _, writeErr := file.Write(configBytes); writeErr != nil {
		return writeErr
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}

	editorCmd := exec.Command(editor, file.Name())
	editorCmd.Stdout = os.Stdout
	editorCmd.Stdin = os.Stdin
	editorCmd.Stderr = os.Stderr

	if cmdErr := editorCmd.Run(); cmdErr != nil {
		return cmdErr
	}

	newConfigB, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(newConfigB, config); err != nil {
		return err
	}

	newConfig, response := c.UpdateConfig(config)
	if response.Error != nil {
		return response.Error
	}

	printer.PrintT("Config updated successfully", newConfig)
	return nil
}

func configResetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")

	if !confirmFlag && len(args) > 0 {
		var confirmResetAll string
		confirmationMsg := fmt.Sprintf(
			"Are you sure you want to reset %s to their default value? (YES/NO): ",
			args[0])
		fmt.Println(confirmationMsg)
		_, _ = fmt.Scanln(&confirmResetAll)
		if confirmResetAll != "YES" {
			fmt.Println("Reset operation aborted")
			return nil
		}
	}

	defaultConfig := &model.Config{}
	defaultConfig.SetDefaults()
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	for _, arg := range args {
		path := parseConfigPath(arg)
		defaultValue, ok := getValue(path, *defaultConfig)
		if !ok {
			return errors.New("invalid key")
		}
		err := resetConfigValue(path, config, defaultValue)
		if err != nil {
			return err
		}
	}
	newConfig, res := c.UpdateConfig(config)
	if res.Error != nil {
		return res.Error
	}

	printer.PrintT("Value/s reset successfully", newConfig)
	return nil
}

func configShowCmdF(c client.Client, _ *cobra.Command, _ []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FormatJSON)
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	printer.Print(config)

	return nil
}

func parseConfigPath(configPath string) []string {
	return strings.Split(configPath, ".")
}
