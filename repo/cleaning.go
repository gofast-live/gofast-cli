package repo

import "os"

func cleaning(projectName string, protocol string, database string, paymentsProvider string, emailProvider string, filesProvider string) error {
	var err error
	if protocol == "HTTP" {
		err = os.RemoveAll(projectName + "/svelte-grpc")
        if err != nil {
            return err
        }
		err = os.RemoveAll(projectName + "/go-grpc")
        if err != nil {
            return err
        }
	} else if protocol == "gRPC" {
		err = os.RemoveAll(projectName + "/svelte-http")
        if err != nil {
            return err
        }
		err = os.RemoveAll(projectName + "/go-http")
        if err != nil {
            return err
        }
	}
	return nil
}
