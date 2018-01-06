package allocation

type SystemPaths struct {
	Local   string
	Runtime string
}

func DefaultSystemPaths() SystemPaths {
	return SystemPaths{
		Local:   dirSystemDLocal,
		Runtime: dirSystemDRuntime,
	}
}
