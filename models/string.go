package models

import "strings"

func SplitCommaSeparatedIPs(ipString string) []string {
	if strings.TrimSpace(ipString) == "" {
		return []string{}
	}

	rawIPs := strings.Split(ipString, ",")

	cleanedIPs := make([]string, 0, len(rawIPs))

	for _, part := range rawIPs {
		trimmedIP := strings.TrimSpace(part)

		if trimmedIP != "" {
			cleanedIPs = append(cleanedIPs, trimmedIP)
		}
	}

	return cleanedIPs
}
