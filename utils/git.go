package utils

import "os/exec"


func CloneAndCheckout(repoURL, commitSHA, dir string) error{
	if err := exec.Command("git","clone", repoURL, dir).Run(); err != nil{
		return err
	}
	cmd := exec.Command("git","checkout", commitSHA)
	cmd.Dir = dir
	return cmd.Run();

}