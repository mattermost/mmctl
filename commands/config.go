package commands

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mmctl/client"
	"github.com/mattermost/mmctl/printer"

	"github.com/spf13/cobra"
)

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
		ConfigResetCmd,
		ConfigShowCmd,
	)
	RootCmd.AddCommand(ConfigCmd)
}

func getValue(path []string, obj interface{}) (interface{}, bool) {
	r := reflect.ValueOf(obj)
	val := r.FieldByName(path[0])

	if val.Kind() == reflect.Invalid {
		return nil, false
	}

	if len(path) == 1 {
		return val.Interface(), true
	} else if val.Kind() == reflect.Struct {
		return getValue(path[1:], val.Interface())
	} else {
		return nil, false
	}
}

func setValue(path []string, obj reflect.Value, newValue interface{}) error {
	var val reflect.Value
	if obj.Kind() == reflect.Struct {
		val = obj.FieldByName(path[0])
	} else {
		val = obj
	}

	if val.Kind() == reflect.Invalid {
		return errors.New("Selected path object is not valid")
	}

	if len(path) > 1 && val.Kind() == reflect.Struct {
		return setValue(path[1:], val, newValue)
	} else if len(path) == 1 {
		if val.Kind() == reflect.Ptr {
			return setValue(path, val.Elem(), newValue)
		} else if val.Kind() == reflect.Struct {
			val.Set(reflect.ValueOf(newValue))
			return nil
		} else if val.Kind() == reflect.Slice {
			if val.Type().Elem().Kind() != reflect.String {
				return errors.New("Unsupported type of slice")
			}
			val.Set(reflect.ValueOf(newValue))
			return nil
		} else {
			switch val.Kind() {
			case reflect.Int:
				v, err := strconv.ParseInt(newValue.(string), 10, 64)
				if err != nil {
					return errors.New("Target value is of type Int and provided value is not")
				}
				val.SetInt(v)
				return nil
			case reflect.String:
				val.SetString(newValue.(string))
				return nil
			case reflect.Bool:
				v, err := strconv.ParseBool(newValue.(string))
				if err != nil {
					return errors.New("Target value is of type Bool and provided value is not")
				}
				val.SetBool(v)
				return nil
			default:
				return errors.New("Target value type is not supported")
			}
		}
	} else {
		return errors.New("Path object type is not supported")
	}
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
		if nv.Elem().Kind() == reflect.Int {
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.Itoa(*newValue.(*int)))
		} else if nv.Elem().Kind() == reflect.Bool {
			return setValue(path, reflect.ValueOf(config).Elem(), strconv.FormatBool(*newValue.(*bool)))
		} else {
			return setValue(path, reflect.ValueOf(config).Elem(), *newValue.(*string))
		}
	} else {
		return setValue(path, reflect.ValueOf(config).Elem(), newValue)
	}
}

func configGetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FORMAT_JSON)

	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path := strings.Split(args[0], ".")
	if val, ok := getValue(path, *config); !ok {
		return errors.New("Invalid key")
	} else {
		printer.Print(val)
	}

	return nil
}

func configSetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	path, err := parseConfigPath(args[0])
	if err != nil {
		return err
	}
	if err := setConfigValue(path, config, args[1:]); err != nil {
		return err
	}
	if _, res := c.UpdateConfig(config); res.Error != nil {
		return res.Error
	}

	printer.Print("Value changed successfully")
	return nil
}

func configResetCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	confirmFlag, _ := cmd.Flags().GetBool("confirm")

	if !confirmFlag && len(args) > 0 {
		var confirmResetAll string
		confirmationMsg := fmt.Sprintf(
			"Are you sure you want to reset %s to their default value? (YES/NO): ",
			args[0])
		printer.Print(confirmationMsg)
		fmt.Scanln(&confirmResetAll)
		if confirmResetAll != "YES" {
			printer.Print("Reset operation aborted")
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
		path, err := parseConfigPath(arg)
		if err != nil {
			return err
		}
		defaultValue, ok := getValue(path, *defaultConfig)
		if !ok {
			return errors.New("Invalid key")
		}
		resetConfigValue(path, config, defaultValue)
	}
	if _, res := c.UpdateConfig(config); res.Error != nil {
		return res.Error
	}

	printer.Print("Value/s reset successfully")
	return nil
}

func configShowCmdF(c client.Client, cmd *cobra.Command, args []string) error {
	printer.SetSingle(true)
	printer.SetFormat(printer.FORMAT_JSON)
	config, response := c.GetConfig()
	if response.Error != nil {
		return response.Error
	}

	printer.Print(config)

	return nil
}

func parseConfigPath(configPath string) ([]string, error) {
	return strings.Split(configPath, "."), nil
}
