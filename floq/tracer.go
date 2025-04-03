package main

type tracer struct
{
	name	string
	path	string
};

func (tra *tracer) frisk_att(al *ast) string {

	for an := al.left;  an != nil;  an = an.next {
		switch an.left.string {
		default:
			return "unknown attribute: " + an.left.string
		}
	}
	return ""
}
