package ping

func Main(args []string) {
	set, params := ParseParams("ping", args)

	if len(params.Target) < 1 {
		set.Usage()
		return
	}

	if params.Classical {
		MainClassical(params)

	} else {
		MainTui(params)
	}
}
