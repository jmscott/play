package main

type sysatt struct {

	name		string
	command_ref	*command
}

func (sa *sysatt) is_uint64() bool {

	if sa.command_ref != nil {
		return sa.command_ref.is_sysatt_uint64(sa.name)
	}
	return false
}

func (sa *sysatt) String() string {

	return sa.command_ref.name + "$" + sa.name
}

func (sa *sysatt) full_name() string {
	if sa.command_ref != nil {
		return sa.command_ref.name + "$" + sa.name
	}
	return ""
}
