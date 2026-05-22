package main

import "github.com/manifoldco/promptui"

var debug int = 1 // Debug 1 for true 0 to disable output

type network_params struct {
	mtu  int
	ip   string
	name string
}

func call_opt(name string) {
	println(
		"Bem vindo " + name + " ao gerenciador de rede!" + "\n" + "==========================================",
	)
	//println("[ 1 ] - Network/IP configuration" + "\n" + "[ 2 ] - DNS" + "\n" + "[ 3 ] - Connection Test" + "\n" + "[ 4 ] - Exit")
}

func get_input() {

	prompt := promptui.Prompt{
		Label: "Digite seu nome",
	}

	select_opt := promptui.Select{
		Label: "Choose an option",
		Items: []string{"[ 1 ] - Network/IP configuration", "[ 2 ] - DNS", "[ 3 ] - Connection Test", "[ 4 ] - Exit"},
	}

	name, err := prompt.Run()

	if err != nil {
		println("Sorry an error has ocurred! %v\n", err)
		return
	}
	//Call menu options
	call_opt(name)

	_, opt, err := select_opt.Run()

	println("You selected: %q\n", opt)
}
