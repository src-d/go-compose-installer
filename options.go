package installer

type InstallOptions struct {
	// NoExec skip the execution of the exec command configured for this command.
	NoExec bool
}

type StartOptions struct {
	// NoExec skip the execution of the exec command configured for this command.
	NoExec bool
}

type StopOptions struct {
	// NoExec skip the execution of the exec command configured for this command.
	NoExec bool
}

type UninstallOptions struct {
	// NoExec skip the execution of the exec command configured for this command.
	NoExec bool
	// Purge remove all the installed images and volumes.
	Purge bool
	// Force force the un-installation even if an installation is not detected.
	Force bool
}
