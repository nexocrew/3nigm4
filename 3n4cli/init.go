//
// 3nigm4 3n4cli package
// Author: Guido Ronchetti <dyst0ni3@gmail.com>
// v1.0 16/06/2016
//

package main

// Third party libs
import (
	"github.com/spf13/cobra"
)

func initIshtm() {
	RootCmd.AddCommand(IshtmCmd)
	setArgument(IshtmCmd, "ishtmeaddress")
	setArgument(IshtmCmd, "ishtmport")
	bindPFlag(IshtmCmd, "ishtmeaddress")
	bindPFlag(IshtmCmd, "ishtmport")

	IshtmCmd.AddCommand(CreateCmd)
	// i/o paths
	setArgument(CreateCmd, "input")
	setArgument(CreateCmd, "output")
	setArgument(CreateCmd, "extension")
	setArgument(CreateCmd, "notify")
	setArgument(CreateCmd, "recipients")
	bindPFlag(CreateCmd, "input")
	bindPFlag(CreateCmd, "output")
	bindPFlag(CreateCmd, "extension")
	bindPFlag(CreateCmd, "notify")
	bindPFlag(CreateCmd, "recipients")

	IshtmCmd.AddCommand(DeleteWillCmd)
	setArgument(DeleteWillCmd, "id")
	setArgument(DeleteWillCmd, "secondary")
	bindPFlag(DeleteWillCmd, "id")
	bindPFlag(DeleteWillCmd, "secondary")

	IshtmCmd.AddCommand(GetCmd)
	setArgument(GetCmd, "output")
	setArgument(GetCmd, "id")
	bindPFlag(GetCmd, "output")
	bindPFlag(GetCmd, "id")

	IshtmCmd.AddCommand(PatchCmd)
	setArgument(PatchCmd, "id")
	setArgument(PatchCmd, "secondary")
	bindPFlag(PatchCmd, "id")
	bindPFlag(PatchCmd, "secondary")
}

func initStorage() {
	RootCmd.AddCommand(StoreCmd)
	// API references
	setArgument(StoreCmd, "storageaddress")
	setArgument(StoreCmd, "storageport")
	// encryption
	setArgument(StoreCmd, "privatekey")
	setArgument(StoreCmd, "publickey")
	setArgument(StoreCmd, "masterkey")
	// working queue setup
	setArgument(StoreCmd, "workerscount")
	setArgument(StoreCmd, "queuesize")
	// i/o paths
	bindPFlag(StoreCmd, "storageaddress")
	bindPFlag(StoreCmd, "storageport")
	bindPFlag(StoreCmd, "privatekey")
	bindPFlag(StoreCmd, "publickey")
	bindPFlag(StoreCmd, "masterkey")
	bindPFlag(StoreCmd, "workerscount")
	bindPFlag(StoreCmd, "queuesize")

	StoreCmd.AddCommand(UploadCmd)
	// encryption
	setArgument(UploadCmd, "destkeys")
	// i/o paths
	setArgument(UploadCmd, "input")
	setArgument(UploadCmd, "referenceout")
	setArgument(UploadCmd, "chunksize")
	setArgument(UploadCmd, "compressed")
	// resource properties
	setArgument(UploadCmd, "timetolive")
	setArgument(UploadCmd, "permission")
	setArgument(UploadCmd, "sharingusers")
	bindPFlag(UploadCmd, "destkeys")
	bindPFlag(UploadCmd, "input")
	bindPFlag(UploadCmd, "referenceout")
	bindPFlag(UploadCmd, "chunksize")
	bindPFlag(UploadCmd, "compressed")
	bindPFlag(UploadCmd, "timetolive")
	bindPFlag(UploadCmd, "permission")
	bindPFlag(UploadCmd, "sharingusers")
	UploadCmd.RunE = upload

	StoreCmd.AddCommand(DownloadCmd)
	// i/o paths
	setArgument(DownloadCmd, "referencein")
	setArgument(DownloadCmd, "output")
	bindPFlag(DownloadCmd, "output")
	bindPFlag(DownloadCmd, "referencein")

	StoreCmd.AddCommand(InfoCmd)
	setArgument(InfoCmd, "referencein")
	bindPFlag(InfoCmd, "referencein")

	StoreCmd.AddCommand(DeleteCmd)
	setArgument(DeleteCmd, "referencein")
	bindPFlag(DeleteCmd, "referencein")
}

func initAuth() {
	RootCmd.AddCommand(LoginCmd)
	setArgument(LoginCmd, "authaddress")
	setArgument(LoginCmd, "authport")
	setArgument(LoginCmd, "username")
	bindPFlag(LoginCmd, "username")
	bindPFlag(LoginCmd, "authaddress")
	bindPFlag(LoginCmd, "authport")
	LoginCmd.RunE = login

	RootCmd.AddCommand(LogoutCmd)
	setArgument(LogoutCmd, "authaddress")
	setArgument(LogoutCmd, "authport")
	bindPFlag(LogoutCmd, "authaddress")
	bindPFlag(LogoutCmd, "authport")
	LogoutCmd.RunE = logout

	RootCmd.AddCommand(CreateUserCmd)
	CreateUserCmd.RunE = createuser
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	setArgument(RootCmd, "verbose")
	bindPFlag(RootCmd, "verbose")
	// Ishtm
	initIshtm()
	// Storage
	initStorage()
	// Authentication
	initAuth()
	// Ping
	RootCmd.AddCommand(PingCmd)
	setArgument(PingCmd, "authaddress")
	setArgument(PingCmd, "authport")
	setArgument(PingCmd, "storageaddress")
	setArgument(PingCmd, "storageport")
	setArgument(PingCmd, "ishtmeaddress")
	setArgument(PingCmd, "ishtmport")
	bindPFlag(PingCmd, "authaddress")
	bindPFlag(PingCmd, "authport")
	bindPFlag(PingCmd, "storageaddress")
	bindPFlag(PingCmd, "storageport")
	bindPFlag(PingCmd, "ishtmeaddress")
	bindPFlag(PingCmd, "ishtmport")

	// Version
	RootCmd.AddCommand(VersionCmd)
}
