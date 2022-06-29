Vim�UnDo� >+f5V�yW���|��T�՞�UU��b�   z           9                       b��    _�                        9    ����                                                                                                                                                                                                                                                                                                                                                             b���    �         �      G	CIDockerTags    = []DockerTagProvider{GitBranch, GitHash, DateGitHash}5��       9                  a                     5�_�                    j   7    ����                                                                                                                                                                                                                                                                                                                            m           j   7       V   :    b���     �   i   j          7// DateGitHash for a time sortable tag in the registry.   func DateGitHash() string {   ]	return fmt.Sprintf("%s_%s", time.Now().UTC().Format("2006_01_02T15_04_05Z07_00"), GitHash())   }5��    i                      �      �               5�_�                    j        ����                                                                                                                                                                                                                                                                                                                            j           j   7       V   :    b���    �   i   j           5��    i                      �                     5�_�                    k        ����                                                                                                                                                                                                                                                                                                                            j           j   7       V   :    b���     �         {      	"fmt"   	"os"�      	   |      
	"strings"   	"time"5��       	                 H                      �                        #                      5�_�                            ����                                                                                                                                                                                                                                                                                                                            h           h   7       V   :    b���    �               z   package repo       import (   		"errors"   	"os"   	"path/filepath"   
	"strings"       	"github.com/go-git/go-git/v5"   	"github.com/hmarr/codeowners"   	"github.com/magefile/mage/sh"   )       $type DockerTagProvider func() string       var (   H	LocalDockerTags = []DockerTagProvider{func() string { return "local" }}   :	CIDockerTags    = []DockerTagProvider{GitBranch, GitHash}   )       func Root() string {   >	path, err := sh.Output("git", "rev-parse", "--show-toplevel")   	if err != nil {   		panic(err)   	}       	return path   }       .func OwnersOf(path string) ([]string, error) {   :	file, err := os.Open(filepath.Join(Root(), "CODEOWNERS"))   	defer file.Close()   	if err != nil {   		return nil, err   	}       +	ruleset, err := codeowners.ParseFile(file)   	if err != nil {   		return nil, err   	}       !	rule, err := ruleset.Match(path)   	if err != nil {   		return nil, err   	}       	var owners []string    	for _, o := range rule.Owners {   "		owners = append(owners, o.Value)   	}       	if len(owners) == 0 {   9		return nil, errors.New("no code owners for given path")   	}       	return owners, nil   }       func Name() (string, error) {   #	repo, err := git.PlainOpen(Root())   	if err != nil {   		return "", err   	}       %	remote, err := repo.Remote("origin")   	if err != nil {   		return "", err   	}       	url := remote.Config().URLs[0]   !	parts := strings.Split(url, "/")   	name := parts[len(parts)-1]       -	return strings.TrimSuffix(name, ".git"), nil   }       func GitSummary() string {   N	summary, err := sh.Output("git", "describe", "--tags", "--dirty", "--always")   	if err != nil {   		panic(err)   	}       	return summary   }       func GitBranch() string {   E	branch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")   	if err != nil {   		panic(err)   	}       	return branch   }       func GitHash() string {   3	hash, err := sh.Output("git", "rev-parse", "HEAD")   	if err != nil {   		panic(err)   	}       	return hash   }       !func DockerImageTags() []string {   	var tags []string       	if os.Getenv("CI") == "" {   		// we're not running in CI   %		for _, p := range LocalDockerTags {   			tags = append(tags, p())   		}       		return tags   	}       	// we're running in CI   !	for _, p := range CIDockerTags {   		tags = append(tags, p())   	}       	return tags   }5�5�_�                            ����                                                                                                                                                                                                                                                                                                                            h           h   7       V   :    b���     �                	"fmt"5��                          !                      5�_�                             ����                                                                                                                                                                                                                                                                                                                            g           g   7       V   :    b��    �                	"time"5��                          C                      5��