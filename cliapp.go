package main

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
)

var debug int = 1 // Debug 1 for true 0 to disable output

type network_params struct {
	mtu  int
	ip   string
	name string
}

var templates = &promptui.SelectTemplates{

	Label: `{{ "⚡" | yellow }} {{ . | bold | cyan }}`,

	Active: `{{ "▸" | green | bold }} {{ . | green | bold }}`,

	Inactive: `  {{ . | white }}`,

	Selected: `{{ "✔" | green | bold }} {{ "Você escolheu:" | white }} {{ . | cyan }}`,
}

const (
	op_one   = "[ 1 ] - Network/IP configuration"
	op_two   = "[ 2 ] - DNS"
	op_three = "[ 3 ] - Connection Test"
	op_four  = "[ 4 ] - Exit"
)

const (
	submenu_one   = "[ 1 ] - Change IP"
	submenu_two   = "[ 2 ] - Manage Routes"
	submenu_three = "[ 3 ] - Return"
)

func call_opt() {
	prompt := promptui.Prompt{
		Label: "Digite seu nome",
	}

	name, err := prompt.Run()

	if err != nil {
		println("An error has ocurred! %p", err)
	}

	println(
		"Bem vindo " + name + " ao gerenciador de rede!" + "\n" + "==========================================",
	)
}

func submenu_opt1(options []string) {

	selectedOpt := promptui.Select{
		Label:     "Network/IP Configuration",
		Items:     options,
		Templates: templates,
	}

	_, opt, err := selectedOpt.Run()

	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) {
			println("\n The operation has been interrupt by user!")
			os.Exit(0)
		}
		println("Sorry, an execption has ocurred: %p\n", err)
	}
	switch opt {
	case submenu_one:
		println("Not Implemented!")
	case submenu_two:
		println("Not Implemented!")
	case submenu_three:
		return
	}
}

func get_input() {

	main_menu := []string{
		op_one,
		op_two,
		op_three,
		op_four,
	}

	submenu := []string{
		submenu_one,
		submenu_two,
		submenu_three,
	}

	select_opt := promptui.Select{
		Label:     "Choose an option",
		Items:     main_menu,
		Templates: templates,
	}

	//Call menu options

	for {
		_, opt, err := select_opt.Run()

		if err != nil {
			if errors.Is(
				err, promptui.ErrInterrupt,
			) {
				println("\n The operation has been interrupt by user!")
				os.Exit(0)
			}
			println("Sorry, an execption has ocurred: %p\n", err)
		}

		//switch to make a decision based on input user

		switch opt {
		case op_one:
			submenu_opt1(submenu)
		case op_two:
			println("Not implemented!")
		case op_three:
			println("Not implemented!")
		case op_four:
			os.Exit(0)
		}

	}
}
