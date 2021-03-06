package pf

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const measurement = "pf"
const pfctlCommand = "pfctl"

type PF struct {
	PfctlCommand string
	PfctlArgs    []string
	UseSudo      bool
	StateTable   []*Entry
	infoFunc     func() (string, error)
}

func (pf *PF) Description() string {
	return "Gather counters from PF"
}

func (pf *PF) SampleConfig() string {
	return `
  ## PF require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run pfctl.
  ## Users must configure sudo to allow telegraf user to run pfctl with no password.
  ## pfctl can be restricted to only list command "pfctl -s info".
  use_sudo = false
`
}

// Gather is the entrypoint for the plugin.
func (pf *PF) Gather(acc telegraf.Accumulator) error {
	if pf.PfctlCommand == "" {
		var err error
		if pf.PfctlCommand, pf.PfctlArgs, err = pf.buildPfctlCmd(); err != nil {
			acc.AddError(fmt.Errorf("Can't construct pfctl commandline: %s", err))
			return nil
		}
	}

	o, err := pf.infoFunc()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	if perr := pf.parsePfctlOutput(o, acc); perr != nil {
		acc.AddError(perr)
	}
	return nil
}

var errParseHeader = fmt.Errorf("Cannot find header in %s output", pfctlCommand)

func errMissingData(tag string) error {
	return fmt.Errorf("struct data for tag \"%s\" not found in %s output", tag, pfctlCommand)
}

type pfctlOutputStanza struct {
	HeaderRE  *regexp.Regexp
	ParseFunc func([]string, telegraf.Accumulator) error
	Found     bool
}

var pfctlOutputStanzas = []*pfctlOutputStanza{
	&pfctlOutputStanza{
		HeaderRE:  regexp.MustCompile("^State Table"),
		ParseFunc: parseStateTable,
	},
}

var anyTableHeaderRE = regexp.MustCompile("^[A-Z]")

func (pf *PF) parsePfctlOutput(pfoutput string, acc telegraf.Accumulator) error {
	scanner := bufio.NewScanner(strings.NewReader(pfoutput))
	for scanner.Scan() {
		line := scanner.Text()
		for _, s := range pfctlOutputStanzas {
			if s.HeaderRE.MatchString(line) {
				var stanzaLines []string
				scanner.Scan()
				line = scanner.Text()
				for !anyTableHeaderRE.MatchString(line) {
					stanzaLines = append(stanzaLines, line)
					scanner.Scan()
					line = scanner.Text()
				}
				if perr := s.ParseFunc(stanzaLines, acc); perr != nil {
					return perr
				}
				s.Found = true
			}
		}
	}
	for _, s := range pfctlOutputStanzas {
		if !s.Found {
			return errParseHeader
		}
	}
	return nil
}

type Entry struct {
	Field      string
	PfctlTitle string
	Value      int64
}

var StateTable = []*Entry{
	&Entry{"entries", "current entries", -1},
	&Entry{"searches", "searches", -1},
	&Entry{"inserts", "inserts", -1},
	&Entry{"removals", "removals", -1},
}

var stateTableRE = regexp.MustCompile(`^  (.*?)\s+(\d+)`)

func parseStateTable(lines []string, acc telegraf.Accumulator) error {
	for _, v := range lines {
		entries := stateTableRE.FindStringSubmatch(v)
		if entries != nil {
			for _, f := range StateTable {
				if f.PfctlTitle == entries[1] {
					var err error
					if f.Value, err = strconv.ParseInt(entries[2], 10, 64); err != nil {
						return err
					}
				}
			}
		}
	}

	fields := make(map[string]interface{})
	for _, v := range StateTable {
		if v.Value == -1 {
			return errMissingData(v.PfctlTitle)
		}
		fields[v.Field] = v.Value
	}

	acc.AddFields(measurement, fields, make(map[string]string))
	return nil
}

func (pf *PF) callPfctl() (string, error) {
	cmd := execCommand(pf.PfctlCommand, pf.PfctlArgs...)
	out, oerr := cmd.Output()
	if oerr != nil {
		ee, ok := oerr.(*exec.ExitError)
		if !ok {
			return string(out), fmt.Errorf("error running %s: %s: (unable to get stderr)", pfctlCommand, oerr)
		}
		return string(out), fmt.Errorf("error running %s: %s: %s", pfctlCommand, oerr, ee.Stderr)
	}
	return string(out), oerr
}

var execLookPath = exec.LookPath
var execCommand = exec.Command

func (pf *PF) buildPfctlCmd() (string, []string, error) {
	cmd, err := execLookPath(pfctlCommand)
	if err != nil {
		return "", nil, fmt.Errorf("can't locate %s: %v", pfctlCommand, err)
	}
	args := []string{"-s", "info"}
	if pf.UseSudo {
		args = append([]string{cmd}, args...)
		cmd, err = execLookPath("sudo")
		if err != nil {
			return "", nil, fmt.Errorf("can't locate sudo: %v", err)
		}
	}
	return cmd, args, nil
}

func init() {
	inputs.Add("pf", func() telegraf.Input {
		pf := new(PF)
		pf.infoFunc = pf.callPfctl
		return pf
	})
}
