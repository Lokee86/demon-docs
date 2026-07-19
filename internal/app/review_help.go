package app

import (
	"fmt"
	"io"
)

func suggestionCommandHelp(out io.Writer, command string) {
	switch command {
	case "declined":
		fmt.Fprintln(out, "usage: ddocs suggestions declined [-h] [FILE]\n\nShow effective declined suggestions, optionally limited to one repository-relative source file. Stale decisions whose evidence fingerprint changed are not effective declines.\n\noptions:\n  -h, --help  show this help message and exit")
	case "log":
		fmt.Fprintln(out, "usage: ddocs suggestions log [-h] [FILE]\n\nShow append-only suggestion decision history, optionally limited to one repository-relative source file.\n\noptions:\n  -h, --help  show this help message and exit")
	case "show":
		fmt.Fprintln(out, "usage: ddocs suggestions show [-h] SUGGESTION\n\nShow one current suggestion or its most recent retained historical form. Candidate numbers, target paths, evidence tier, score, and decline state are included when available.\n\noptions:\n  -h, --help  show this help message and exit")
	case "select":
		fmt.Fprintln(out, "usage: ddocs suggestions select [-h] SUGGESTION [CANDIDATE]\n\nApply one current candidate through the normal hash-guarded repair path. CANDIDATE may be the displayed number or target path and may be omitted only when exactly one candidate exists. Declined or blocked suggestions must be reconsidered or unblocked first.\n\noptions:\n  -h, --help  show this help message and exit")
	case "decline":
		fmt.Fprintln(out, "usage: ddocs suggestions decline [-h] SUGGESTION [CANDIDATE] [--reason TEXT]\n\nDecline one candidate or, when CANDIDATE is omitted, the whole suggestion. The decision remains effective while the relationship and evidence fingerprint remain unchanged.\n\noptions:\n  -h, --help     show this help message and exit\n  --reason TEXT  record an optional review reason")
	case "reconsider":
		fmt.Fprintln(out, "usage: ddocs suggestions reconsider [-h] SUGGESTION\n\nClear effective candidate-level and issue-level declines for one current or historical suggestion so it can be reviewed again.\n\noptions:\n  -h, --help  show this help message and exit")
	}
}

func changeCommandHelp(out io.Writer, command string) {
	switch command {
	case "related":
		fmt.Fprintln(out, "usage: ddocs changes related [-h] FILE\n\nShow applied changes caused by or targeting FILE. Current and retained historical paths for tracked files are recognized.\n\noptions:\n  -h, --help  show this help message and exit")
	case "show":
		fmt.Fprintln(out, "usage: ddocs changes show [-h] CHANGE\n\nShow one applied change, its run, hashes, individual repair identifiers, and retained unified diff when available.\n\noptions:\n  -h, --help  show this help message and exit")
	case "log":
		fmt.Fprintln(out, "usage: ddocs changes log [-h] [FILE]\n\nShow append-only applied-change and repair-control history, optionally limited to one tracked source file.\n\noptions:\n  -h, --help  show this help message and exit")
	case "undo":
		fmt.Fprintln(out, "usage: ddocs changes undo [-h] CHANGE [--repair REPAIR] [--block] [--reason TEXT]\n\nUndo an eligible recorded change, or one named repair inside it. Undo is refused when the current file no longer matches the recorded after hash or the configured undo limits have expired.\n\noptions:\n  -h, --help     show this help message and exit\n  --repair ID    undo only the displayed repair ID\n  --block        also block the undone repair fingerprint\n  --reason TEXT  record an optional block reason")
	case "undo-run":
		fmt.Fprintln(out, "usage: ddocs changes undo-run [-h] RUN [--block] [--reason TEXT]\n\nUndo every eligible change in one reconciliation run after all files pass hash and undo-limit preflight. No partial run undo is applied when preflight fails.\n\noptions:\n  -h, --help     show this help message and exit\n  --block        also block every undone repair fingerprint\n  --reason TEXT  record an optional block reason")
	case "block":
		fmt.Fprintln(out, "usage: ddocs changes block [-h] CHANGE [--repair REPAIR] [--reason TEXT]\n\nBlock every repair in one recorded change, or one displayed repair ID, from being applied again while its exact relationship fingerprint remains current.\n\noptions:\n  -h, --help     show this help message and exit\n  --repair ID    block only the displayed repair ID\n  --reason TEXT  record an optional review reason")
	case "unblock":
		fmt.Fprintln(out, "usage: ddocs changes unblock [-h] CHANGE [--repair REPAIR]\n\nRemove active repair blocks for every repair in one recorded change, or one displayed repair ID.\n\noptions:\n  -h, --help   show this help message and exit\n  --repair ID  unblock only the displayed repair ID")
	}
}
