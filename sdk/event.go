package sdk

// Event is returned by the SDK method diff.
// IMPORTANT NOTE:
// Pay close attention to deletedfileid field set in metadata returned from either modifyfile
// or createfile when one file is atomically replaced with another one.
// Clients are advised to ignore events that they don't understand (as opposed to issuing
// errors). For shares, a share object is provided.
// https://docs.pcloud.com/structures/event.html
type Event string

const (
	// Reset event: client should reset it's state to empty root directory.
	Reset Event = "reset"

	// CreateFolder event: folder is created, metadata is provided.
	CreateFolder Event = "createfolder"

	// DeleteFolder event: folder is deleted, metadata is provided.
	DeleteFolder Event = "deletefolder"

	// ModifyFolder event: folder is modified, metadata is provided.
	ModifyFolder Event = "modifyfolder"

	// CreateFile event: file is created, metadata is provided.
	CreateFile Event = "createfile"

	// ModifyFile event: file data is modified, metadata is provided (normally modifytime,
	// size and hash are changed).
	ModifyFile Event = "modifyfile"

	// DeleteFile event: file is deleted, metadata is provided.
	DeleteFile Event = "deletefile"

	// RequestShareIn event: incoming share, share is provided.
	RequestShareIn Event = "requestsharein"

	// AcceptedShareIn event: you have accepted a share request (potentially on another device),
	// useful to decrement the counter of pending requests. share is provided.
	// It is guaranteed that you receive createfolder for the folderid (and all the contents of
	// the folder) of the share before you receive acceptedshare, so it is safe to assume that
	// you will be able to find folderid in the local state.
	AcceptedShareIn Event = "acceptedsharein"

	// DeclinedShareIn event: you have declined a share request, share is provided (this is
	// delivered to the declining user, not to the sending one).
	DeclinedShareIn Event = "declinedsharein"

	// DeclinedShareOut event: same as above, but delivered to the user that is sharing the
	// folder.
	DeclinedShareOut Event = "declinedshareout"

	// CancelledShareIn event: the sender of a share request cancelled the share request.
	CancelledShareIn Event = "cancelledsharein"

	// RemovedShareIn event: your incoming share is removed (either by you or the other user).
	RemovedShareIn Event = "removedsharein"

	// ModifiedShareIn event: your incoming share in is modified (permissions changed).
	ModifiedShareIn Event = "modifiedsharein"

	// ModifyUserInfo event: user's information is modified, includes userinfo object with
	// the following fields:
	//    - userid
	//    - premium
	//    - premiumexpires (if premium is true)
	//    - language
	//    - email
	//    - emailverified
	//    - quota
	//    - usedquota.
	// Every user is guaranteed to have one such event in it's full state diff.
	ModifyUserInfo Event = "modifyuserinfo"
)
