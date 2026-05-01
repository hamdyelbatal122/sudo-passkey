package execx

import "os/exec"

func OpenBrowser(url string) error {
	commands := [][]string{
		{"xdg-open", url},
		{"gio", "open", url},
		{"open", url},
	}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd[0]); err != nil {
			continue
		}
		if err := exec.Command(cmd[0], cmd[1:]...).Start(); err == nil {
			return nil
		}
	}

	return nil
}
