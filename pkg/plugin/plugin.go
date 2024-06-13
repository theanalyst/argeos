package plugin

type Plugin interface {
	Name() string // Name of the plugin
	Run() error   // Run the plugin
}
