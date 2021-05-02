package reconciler

const (
	P2PPortName          = "p2p"
	P2PPortNumber uint16 = 18080

	RestrictedPortName          = "restricted"
	RestrictedPortNumber uint16 = 18089

	MonerodContainerName      = "monerod"
	MonerodContainerImage     = "index.docker.io/utxobr/monerod@sha256:19ba5793c00375e7115469de9c14fcad928df5867c76ab5de099e83f646e175d"
	MonerodContainerProbePath = "/get_info"
	MonerodContainerProbePort = RestrictedPortName

	MonerodDataVolumeName      = "data"
	MonerodDataVolumeMountPath = "/data"

	MonerodConfigVolumeName      = "monerod-conf"
	MonerodConfigVolumeMountPath = "/monerod-conf"
)
