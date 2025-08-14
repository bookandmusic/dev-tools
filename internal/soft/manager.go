package soft

type SoftManager interface {
	Install() error
	Uninstall() error
	Update() error
}
