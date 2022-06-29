Vim�UnDo� \ɥj��x@��'�jI���IX�o���X𖵸��   _   &// // Dev runs the application locally   A                           b�޲    _�                            ����                                                                                                                                                                                                                                                                                                                                                 V       b��t     �                !const kubeNamespace = "clubhouse"       func init() {       +	targets.WithRequiredConfig(targets.Config{   .		ServiceName:         "dex-timeline-emailer",   *		AppName:             "timeline-emailer",   .		DockerImage:         "dex/timeline-emailer",   %		KubernetesNamespace: kubeNamespace,   	})   }5��                          �                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                 V       b��t     �                 5��                          �                      5�_�                            ����                                                                                                                                                                                                                                                                                                                                                 V       b��u    �         `          1	return sh.RunV(targets.GoBin, "mod", "download")�         a      func Test() error {    �   '   )   b      0func SendTestEmails(ctx context.Context) error {    �   E   G   c      %func Dev(ctx context.Context) error {    5��    E   %                                        �    '   0                  �                     �                         �                     �                          4                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                  V       b���     �                	//mage:import   7	"github.com/utilitywarehouse/dex-mage-targets/targets"5��                          �       G               5�_�                            ����                                                                                                                                                                                                                                                                                                                                                  V       b���     �                 5��                          �                      5�_�                            ����                                                                                                                                                                                                                                                                                                                                                  V       b���    �         \      )5��                          �               8       5�_�                    ]        ����                                                                                                                                                                                                                                                                                                                                       ]           V        b���     �              N   func Install() error {   1	return sh.RunV(targets.GoBin, "mod", "download")   }       I// Test runs all go tests in the current directory and all subdirectories   func Test() error {   @	if err := sh.RunV(targets.ComposeBin, "up", "-d"); err != nil {   		return err   	}   	// Give it a moment to boot   	time.Sleep(2 * time.Second)       	envVars := map[string]string{   %		"TEST_LOCAL_DB_DRIVER": "postgres",   X		"TEST_LOCAL_DB_DSN":    "postgresql://root@localhost:26258/defaultdb?sslmode=disable",   	}   L	return sh.RunWithV(envVars, targets.GoBin, "test", "-v", "-cover", "./...")   }       L// SendTestEmails generates test e-mails and sends them to the given address   0func SendTestEmails(ctx context.Context) error {   =	targets.PortForwardingTargets = []targets.PortForwardTarget{   z		{Context: "dev-aws", Namespace: "digital-support", Entity: "deployment/click-uw-je", PortLocal: 8989, PortRemote: 8989},   k		{Context: "dev-merit", Namespace: "unicom", Entity: "deployment/api", PortLocal: 8090, PortRemote: 8090},   	}       '	ctx, cancel := context.WithCancel(ctx)   	defer cancel()       	go targets.PortForward(ctx)   	// Give it a moment to boot   	time.Sleep(2 * time.Second)       	defer func() {   5		if err := sh.RunV("pkill", "kubectl"); err != nil {   P			dexlog.NewLogger(true).Error.Println("failed to kill kubectl processes", err)   		}   	}()       	return sh.RunV(   		targets.GoBin,   		"run",   !		"cmd/timeline-emailer/main.go",   -		"cmd/timeline-emailer/send_test_emails.go",   		"send-test-emails",   	)   }       #// Dev runs the application locally   %func Dev(ctx context.Context) error {   =	targets.PortForwardingTargets = []targets.PortForwardTarget{   r		{Context: "dev-aws", Namespace: kubeNamespace, Entity: "deployment/proximo", PortLocal: 6868, PortRemote: 6868},   z		{Context: "dev-merit", Namespace: "ordering-platform", Entity: "deployment/proximo", PortLocal: 6867, PortRemote: 6868},   y		{Context: "dev-merit", Namespace: "account-platform", Entity: "deployment/proximo", PortLocal: 6869, PortRemote: 6868},   z		{Context: "dev-aws", Namespace: "digital-support", Entity: "deployment/click-uw-je", PortLocal: 8989, PortRemote: 8989},   k		{Context: "dev-merit", Namespace: "unicom", Entity: "deployment/api", PortLocal: 8090, PortRemote: 8090},   	}       '	ctx, cancel := context.WithCancel(ctx)   	defer cancel()       	go targets.PortForward(ctx)       @	if err := sh.RunV(targets.ComposeBin, "up", "-d"); err != nil {   		return err   	}       	// Give it a moment to boot   	time.Sleep(2 * time.Second)       	defer func() {   5		if err := sh.RunV("pkill", "kubectl"); err != nil {   P			dexlog.NewLogger(true).Error.Println("failed to kill kubectl processes", err)   		}   	}()       q	return sh.RunV(targets.GoBin, "run", "cmd/timeline-emailer/main.go", "cmd/timeline-emailer/send_test_emails.go")   }5��           N       N             k
      G      5�_�      	                      ����                                                                                                                                                                                                                                                                                                                                                  V        b���    �         ]      import (   
	"context"   	"time"       	"github.com/magefile/mage/sh"   0	dexlog "github.com/utilitywarehouse/dex-go-log"   7	"github.com/utilitywarehouse/dex-mage-targets/targets"   )5��                         .       �       �       5�_�      
           	   A        ����                                                                                                                                                                                                                                                                                                                            ]           A           V        b���     �   @              (// func Dev(ctx context.Context) error {   @// 	targets.PortForwardingTargets = []targets.PortForwardTarget{   u// 		{Context: "dev-aws", Namespace: kubeNamespace, Entity: "deployment/proximo", PortLocal: 6868, PortRemote: 6868},   }// 		{Context: "dev-merit", Namespace: "ordering-platform", Entity: "deployment/proximo", PortLocal: 6867, PortRemote: 6868},   |// 		{Context: "dev-merit", Namespace: "account-platform", Entity: "deployment/proximo", PortLocal: 6869, PortRemote: 6868},   }// 		{Context: "dev-aws", Namespace: "digital-support", Entity: "deployment/click-uw-je", PortLocal: 8989, PortRemote: 8989},   n// 		{Context: "dev-merit", Namespace: "unicom", Entity: "deployment/api", PortLocal: 8090, PortRemote: 8090},   // 	}   //   *// 	ctx, cancel := context.WithCancel(ctx)   // 	defer cancel()   //   // 	go targets.PortForward(ctx)   //   C// 	if err := sh.RunV(targets.ComposeBin, "up", "-d"); err != nil {   // 		return err   // 	}   //   // 	// Give it a moment to boot   // 	time.Sleep(2 * time.Second)   //   // 	defer func() {   8// 		if err := sh.RunV("pkill", "kubectl"); err != nil {   S// 			dexlog.NewLogger(true).Error.Println("failed to kill kubectl processes", err)   // 		}   // 	}()   //   t// 	return sh.RunV(targets.GoBin, "run", "cmd/timeline-emailer/main.go", "cmd/timeline-emailer/send_test_emails.go")   // }5��    @                     j            �      5�_�   	              
           ����                                                                                                                                                                                                                                                                                                                                                  V        b��    �         ]      // import (   // 	"context"   
// 	"time"   //   !// 	"github.com/magefile/mage/sh"   3// 	dexlog "github.com/utilitywarehouse/dex-go-log"   :// 	"github.com/utilitywarehouse/dex-mage-targets/targets"   // )5��                         .       �       �       5�_�   
                 !       ����                                                                                                                                                                                                                                                                                                                                      !          V       b��3     �      "   ]      // func Test() error {   C// 	if err := sh.RunV(targets.ComposeBin, "up", "-d"); err != nil {   // 		return err   // 	}   // 	// Give it a moment to boot   // 	time.Sleep(2 * time.Second)   //   !// 	envVars := map[string]string{   (// 		"TEST_LOCAL_DB_DRIVER": "postgres",   [// 		"TEST_LOCAL_DB_DSN":    "postgresql://root@localhost:26258/defaultdb?sslmode=disable",   // 	}   O// 	return sh.RunWithV(envVars, targets.GoBin, "test", "-v", "-cover", "./...")   // }5��                         �      �      �      5�_�                    "       ����                                                                                                                                                                                                                                                                                                                                      !          V       b��*    �   !   $   ]      //5��    !                      A                     5�_�                    %       ����                                                                                                                                                                                                                                                                                                                            @          %          V       b���    �   $   A   ^      3// func SendTestEmails(ctx context.Context) error {   @// 	targets.PortForwardingTargets = []targets.PortForwardTarget{   }// 		{Context: "dev-aws", Namespace: "digital-support", Entity: "deployment/click-uw-je", PortLocal: 8989, PortRemote: 8989},   n// 		{Context: "dev-merit", Namespace: "unicom", Entity: "deployment/api", PortLocal: 8090, PortRemote: 8090},   // 	}   //   *// 	ctx, cancel := context.WithCancel(ctx)   // 	defer cancel()   //   // 	go targets.PortForward(ctx)   // 	// Give it a moment to boot   // 	time.Sleep(2 * time.Second)   //   // 	defer func() {   8// 		if err := sh.RunV("pkill", "kubectl"); err != nil {   S// 			dexlog.NewLogger(true).Error.Println("failed to kill kubectl processes", err)   // 		}   // 	}()   //   // 	return sh.RunV(   // 		targets.GoBin,   // 		"run",   $// 		"cmd/timeline-emailer/main.go",   0// 		"cmd/timeline-emailer/send_test_emails.go",   // 		"send-test-emails",   // 	)   // }   //5��    $                     �      r      #      5�_�                           ����                                                                                                                                                                                                                                                                                                                                                V       b���     �         ^      // func Install() error {   4// 	return sh.RunV(targets.GoBin, "mod", "download")   // }5��                               T       K       5�_�                           ����                                                                                                                                                                                                                                                                                                                                                V       b���    �         ^      //5��                          X                     5�_�                           ����                                                                                                                                                                                                                                                                                                                                                V       b�ޏ   	 �                //5��                          Y                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                V       b�ޒ     �         ^      L// // Test runs all go tests in the current directory and all subdirectories5��                          Y                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                V       b�ޒ     �         ^      K/ // Test runs all go tests in the current directory and all subdirectories5��                          Y                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                V       b�ޒ     �         ^      J // Test runs all go tests in the current directory and all subdirectories5��                          Y                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                V       b�ޓ     �         ^      I// Test runs all go tests in the current directory and all subdirectories5��                          Y                     5�_�                            ����                                                                                                                                                                                                                                                                                                                                                V       b�ޔ     �         ^      H/ Test runs all go tests in the current directory and all subdirectories5��                          Y                     5�_�                    #        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޖ     �   "   #          //5��    "                      4                     5�_�                    #        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޗ     �   "   $   ]      O// // SendTestEmails generates test e-mails and sends them to the given address5��    "                      4                     5�_�                    #        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޗ     �   "   $   ]      N/ // SendTestEmails generates test e-mails and sends them to the given address5��    "                      4                     5�_�                    #        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޗ   
 �   "   $   ]      M // SendTestEmails generates test e-mails and sends them to the given address5��    "                      4                     5�_�                    ?        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޮ     �   ?   A   ^       �   @   A   ^    �   ?   A   ]    5��    ?                      �                     �    ?                   !   �              !       5�_�                    A        ����                                                                                                                                                                                                                                                                                                                                                V       b�ް     �   @   B   ^      &// // Dev runs the application locally5��    @                      �                     5�_�                    A        ����                                                                                                                                                                                                                                                                                                                                                V       b�ް     �   @   B   ^      %/ // Dev runs the application locally5��    @                      �                     5�_�                    A        ����                                                                                                                                                                                                                                                                                                                                                V       b�ް     �   @   B   ^      $ // Dev runs the application locally5��    @                      �                     5�_�                     A        ����                                                                                                                                                                                                                                                                                                                                                V       b�ޱ    �   @   C   ^      #// Dev runs the application locally5��    @                      �                     5��