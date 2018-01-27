package provision

import (
	"fmt"
	"github.com/akaspin/soil/agent/allocation"
	"github.com/coreos/go-systemd/dbus"
	"os"
)

const (
	phaseDestroyCommand = iota // execute unit commands on destroy
	phaseDestroyUnits          // Destroy units from filesystem
	phaseDeployFS              // Write units to filesystem
	phaseDeployPerm            // Enable or disable units
	phaseDeployCommand         // Execute create/modify unit commands
	phaseDestroyBlobs          // Destroy blobs from filesystem
)

// Instruction represents one atomic instruction bounded to specific phase
type Instruction interface {
	Phase() int
	Execute(conn *dbus.Conn) (err error)
	String() string
}

type baseUnitInstruction struct {
	phase    int
	explain  string
	unitFile allocation.UnitFile
}

func newBaseInstruction(phase int, explain string, unitFile allocation.UnitFile) *baseUnitInstruction {
	return &baseUnitInstruction{
		phase:    phase,
		explain:  explain,
		unitFile: unitFile,
	}
}

func (i *baseUnitInstruction) Phase() int {
	return i.phase
}

func (i *baseUnitInstruction) String() string {
	return fmt.Sprintf("%d:%s:%s", i.phase, i.explain, i.unitFile.Path)
}

// WriteUnitInstruction writes unitFile to filesystem and runs daemon reload.
type WriteUnitInstruction struct {
	*baseUnitInstruction
}

func NewWriteUnitInstruction(unitFile allocation.UnitFile) *WriteUnitInstruction {
	return &WriteUnitInstruction{
		newBaseInstruction(phaseDeployFS, "write-unit", unitFile),
	}
}

func (i *WriteUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	if err = i.unitFile.Write(); err != nil {
		return err
	}
	return conn.Reload()
}

// DeleteUnitInstruction disables and removes unit from systemd
type DeleteUnitInstruction struct {
	*baseUnitInstruction
}

func NewDeleteUnitInstruction(unitFile allocation.UnitFile) *DeleteUnitInstruction {
	return &DeleteUnitInstruction{newBaseInstruction(phaseDestroyUnits, "delete-unit", unitFile)}
}

func (i *DeleteUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	conn.DisableUnitFiles([]string{i.unitFile.UnitName()}, i.unitFile.IsRuntime())
	if err = os.Remove(i.unitFile.Path); err != nil {
		return err
	}
	return conn.Reload()
}

type EnableUnitInstruction struct {
	*baseUnitInstruction
}

func NewEnableUnitInstruction(unitFile allocation.UnitFile) *EnableUnitInstruction {
	return &EnableUnitInstruction{newBaseInstruction(phaseDeployPerm, "enable-unit", unitFile)}
}

func (i *EnableUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	_, _, err = conn.EnableUnitFiles([]string{i.unitFile.Path}, i.unitFile.IsRuntime(), false)
	return err
}

type DisableUnitInstruction struct {
	*baseUnitInstruction
}

func NewDisableUnitInstruction(unitFile allocation.UnitFile) *DisableUnitInstruction {
	return &DisableUnitInstruction{newBaseInstruction(phaseDeployPerm, "disable-unit", unitFile)}
}

func (i *DisableUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	_, err = conn.DisableUnitFiles([]string{i.unitFile.UnitName()}, i.unitFile.IsRuntime())
	return err
}

type CommandInstruction struct {
	*baseUnitInstruction
	command string
}

func NewCommandInstruction(phase int, unitFile allocation.UnitFile, command string) *CommandInstruction {
	return &CommandInstruction{
		baseUnitInstruction: newBaseInstruction(phase, command, unitFile),
		command:             command,
	}
}

func (i *CommandInstruction) Execute(conn *dbus.Conn) (err error) {
	ch := make(chan string)
	switch i.command {
	case "start":
		_, err = conn.StartUnit(i.unitFile.UnitName(), "replace", ch)
	case "restart":
		_, err = conn.RestartUnit(i.unitFile.UnitName(), "replace", ch)
	case "stop":
		_, err = conn.StopUnit(i.unitFile.UnitName(), "replace", ch)
	case "reload":
		_, err = conn.RestartUnit(i.unitFile.UnitName(), "replace", ch)
	case "try-restart":
		_, err = conn.TryRestartUnit(i.unitFile.UnitName(), "replace", ch)
	case "reload-or-restart":
		_, err = conn.ReloadOrRestartUnit(i.unitFile.UnitName(), "replace", ch)
	case "reload-or-try-restart":
		_, err = conn.ReloadOrTryRestartUnit(i.unitFile.UnitName(), "replace", ch)
	default:
		err = fmt.Errorf("unknown systemd command %s", i.command)
	}
	if err != nil {
		return err
	}
	<-ch
	return nil
}

type baseBlobInstruction struct {
	phase   int
	explain string
	blob    *allocation.Blob
}

func (i *baseBlobInstruction) Phase() int {
	return i.phase
}

func (i *baseBlobInstruction) String() string {
	return fmt.Sprintf("%d:%s:%s", i.phase, i.explain, i.blob.Name)
}

type WriteBlobInstruction struct {
	*baseBlobInstruction
}

func NewWriteBlobInstruction(phase int, blob *allocation.Blob) (i *WriteBlobInstruction) {
	return &WriteBlobInstruction{
		&baseBlobInstruction{
			phase:   phase,
			explain: "write-blob",
			blob:    blob,
		},
	}
}

func (i *WriteBlobInstruction) Execute(conn *dbus.Conn) (err error) {
	return i.baseBlobInstruction.blob.Write()
}

type DestroyBlobInstruction struct {
	*baseBlobInstruction
}

func NewDestroyBlobInstruction(phase int, blob *allocation.Blob) (i *DestroyBlobInstruction) {
	return &DestroyBlobInstruction{
		&baseBlobInstruction{
			phase:   phase,
			explain: "delete-blob",
			blob:    blob,
		},
	}
}

func (i *DestroyBlobInstruction) Execute(conn *dbus.Conn) (err error) {
	return os.Remove(i.blob.Name)
}
