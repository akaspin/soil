package scheduler

import (
	"fmt"
	"github.com/coreos/go-systemd/dbus"
	"os"
)

const (
	phaseDestroyCommand = iota // execute unit commands on destroy
	phaseDestroyFS             // Destroy units and blobs
	phaseDeployFS              // Write units and blobs
	phaseDeployPerm
	phaseDeployCommand
)

type Instruction interface {
	Phase() int
	Execute(conn *dbus.Conn) (err error)
}

type baseInstruction struct {
	phase    int
	unitFile *AllocationFile
}

func newBaseInstruction(phase int, unitFile *AllocationFile) *baseInstruction {
	return &baseInstruction{
		phase:    phase,
		unitFile: unitFile,
	}
}

func (i *baseInstruction) Phase() int {
	return i.phase
}

// WriteUnitInstruction writes unitFile to filesystem and runs daemon reload.
type WriteUnitInstruction struct {
	*baseInstruction
}

func NewWriteUnitInstruction(unitFile *AllocationFile) *WriteUnitInstruction {
	return &WriteUnitInstruction{
		newBaseInstruction(phaseDeployFS, unitFile),
	}
}

func (i *WriteUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	if err = i.unitFile.Write(); err != nil {
		return
	}
	err = conn.Reload()
	return
}

func (i *WriteUnitInstruction) String() (res string) {
	res = fmt.Sprintf("%d:write:%s", i.phase, i.unitFile.Path)
	return
}

// DeleteUnitInstruction disables and removes unit from systemd
type DeleteUnitInstruction struct {
	*baseInstruction
}

func NewDeleteUnitInstruction(unitFile *AllocationFile) *DeleteUnitInstruction {
	return &DeleteUnitInstruction{newBaseInstruction(phaseDestroyFS, unitFile)}
}

func (i *DeleteUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	conn.DisableUnitFiles([]string{i.unitFile.UnitName()}, i.unitFile.IsRuntime())
	if err = os.Remove(i.unitFile.Path); err != nil {
		return
	}
	err = conn.Reload()
	return
}

func (i *DeleteUnitInstruction) String() (res string) {
	res = fmt.Sprintf("%d:remove:%s", i.phase, i.unitFile.Path)
	return
}

type EnableUnitInstruction struct {
	*baseInstruction
}

func NewEnableUnitInstruction(unitFile *AllocationFile) *EnableUnitInstruction {
	return &EnableUnitInstruction{newBaseInstruction(phaseDeployPerm, unitFile)}
}

func (i *EnableUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	_, _, err = conn.EnableUnitFiles([]string{i.unitFile.Path}, i.unitFile.IsRuntime(), false)
	return
}

func (i *EnableUnitInstruction) String() (res string) {
	res = fmt.Sprintf("%d:enable:%s", i.phase, i.unitFile.Path)
	return
}

type DisableUnitInstruction struct {
	*baseInstruction
}

func NewDisableUnitInstruction(unitFile *AllocationFile) *DisableUnitInstruction {
	return &DisableUnitInstruction{newBaseInstruction(phaseDeployPerm, unitFile)}
}

func (i *DisableUnitInstruction) Execute(conn *dbus.Conn) (err error) {
	_, err = conn.DisableUnitFiles([]string{i.unitFile.UnitName()}, i.unitFile.IsRuntime())
	return
}

func (i *DisableUnitInstruction) String() (res string) {
	res = fmt.Sprintf("%d:disable:%s", i.phase, i.unitFile.Path)
	return
}

type CommandInstruction struct {
	*baseInstruction
	command string
}

func NewCommandInstruction(phase int, unitFile *AllocationFile, command string) *CommandInstruction {
	return &CommandInstruction{
		baseInstruction: newBaseInstruction(phase, unitFile),
		command:         command,
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
		return
	}
	<-ch
	return
}

func (i *CommandInstruction) String() (res string) {
	res = fmt.Sprintf("%d:%s:%s", i.phase, i.command, i.unitFile.Path)
	return
}

type baseBlobInstruction struct {
	phase int
	blob  *AllocationBlob
}

func (i *baseBlobInstruction) Phase() int {
	return i.phase
}

type WriteBlobInstruction struct {
	*baseBlobInstruction
}

func NewWriteBlobInstruction(phase int, blob *AllocationBlob) (i *WriteBlobInstruction) {
	i = &WriteBlobInstruction{
		&baseBlobInstruction{
			phase: phase,
			blob:  blob,
		},
	}
	return
}

func (i *WriteBlobInstruction) Execute(conn *dbus.Conn) (err error) {
	err = i.baseBlobInstruction.blob.Write()
	return
}

func (i *WriteBlobInstruction) String() (res string) {
	res = fmt.Sprintf("%d:blob-write:%s", i.phase, i.blob.Name)
	return
}

type DestroyBlobInstruction struct {
	*baseBlobInstruction
}

func NewDestroyBlobInstruction(phase int, blob *AllocationBlob) (i *DestroyBlobInstruction) {
	i = &DestroyBlobInstruction{
		&baseBlobInstruction{
			phase: phase,
			blob:  blob,
		},
	}
	return
}

func (i *DestroyBlobInstruction) Execute(conn *dbus.Conn) (err error) {
	err = os.Remove(i.blob.Name)
	return
}

func (i *DestroyBlobInstruction) String() (res string) {
	res = fmt.Sprintf("%d:blob-destroy:%s", i.phase, i.blob.Name)
	return
}
