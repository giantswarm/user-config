package userconfig

type ComponentNames []ComponentName

func (cns ComponentNames) Contain(name ComponentName) bool {
	for _, cn := range cns {
		if cn.Equals(name) {
			return true
		}
	}

	return false
}
