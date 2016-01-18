//A package where information relevant across RDPG can be stored.
package globals

/* By the nature of this package, many files may need to import this package, and
   therefore imports of other packages from RDPG from within this package should
   be extremely minimal. Basically, only log. */
import (
	"fmt"
	"os"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/starkandwayne/rdpgd/log"
)

var (
	MyIP           string
	ServiceRole    string //Gets set in main->parseArgs()
	ClusterService string
	ClusterID      string
)

func init() {
	// Set MyIP variable
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("config.init() consulapi.NewClient()! %s", err))
	} else {
		agent := client.Agent()
		info, err := agent.Self()
		if err != nil {
			log.Error(fmt.Sprintf("config.init() agent.Self()! %s", err))
		} else {
			MyIP = info["Config"]["AdvertiseAddr"].(string)
		}
	}

	ClusterService = os.Getenv("RDPGD_CLUSTER_SERVICE")

	//Set up the ClusterID
	MatrixName := os.Getenv(`RDPGD_MATRIX`)
	MatrixNameSplit := strings.SplitAfterN(MatrixName, `-`, -1)
	MatrixColumn := os.Getenv(`RDPGD_MATRIX_COLUMN`)
	ClusterID = os.Getenv("RDPGD_CLUSTER")
	if ClusterID == "" {
		for i := 0; i < len(MatrixNameSplit)-1; i++ {
			ClusterID = ClusterID + MatrixNameSplit[i]
		}
		ClusterID = ClusterID + "c" + MatrixColumn
	}
}
