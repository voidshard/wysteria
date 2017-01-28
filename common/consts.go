package wysteria_common

const (
	// Message subjects for talking to the server(s)
	// Used to determine what the Message.Data payload should contain
	MSG_CREATE_COLLECTION   = "C_COL"
	MSG_CREATE_ITEM         = "C_ITEM"
	MSG_CREATE_VERSION      = "C_VER"
	MSG_CREATE_RESOURCE = "C_RES"
	MSG_CREATE_LINK         = "C_LINK"

	MSG_DELETE_COLLECTION   = "D_COL"
	MSG_DELETE_ITEM         = "D_ITEM"
	MSG_DELETE_VERSION      = "D_VER"
	MSG_DELETE_RESOURCE = "D_RES"

	MSG_FIND_COLLECTION   = "F_COL"
	MSG_FIND_ITEM         = "F_ITM"
	MSG_FIND_VERSION      = "F_VER"
	MSG_FIND_RESOURCE = "F_FRS"
	MSG_FIND_LINK         = "F_LNK"

	MSG_FIND_HIGHEST_VERSION = "H_VER"

	MSG_UPDATE_ITEM    = "U_ITM"
	MSG_UPDATE_VERSION = "U_VER"

	// server adds to front of err messages sent back
	WYSTERIA_SERVER_ERR = "WYSTERIA_SERVER_ERR:"

	// server replies with to say "Done, nothing went wrong"
)

var (
	// generic server acknowledgement
	WYSTERIA_SERVER_ACK = []byte("WYSTERIA_SERVER_ACK")
)
