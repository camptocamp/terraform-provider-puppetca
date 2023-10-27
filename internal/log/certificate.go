package log

func CertificateFields(nodeName string, environment string) map[string]any {
	return map[string]any{
		"node_name":   nodeName,
		"environment": environment,
	}
}
