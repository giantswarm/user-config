package userconfig

type ComponentNames []ComponentName

func (cns ComponentNames) Contains(name ComponentName) bool {
	for _, cn := range cns {
		if cn.Equals(name) {
			return true
		}
	}

	return false
}
