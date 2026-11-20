#!/usr/bin/env bash

go get github.com/bwmarrin/discordgo
go mod edit -replace github.com/bwmarrin/discordgo=github.com/0x00-sys/discordgo@master
go mod tidy
