package sdk

// https://github.com/pcloudcom/pclouddoc/blob/master/errors.txt
// https://docs.pcloud.com/errors/
const (
	// ErrLoginRequired is returned when log in required.
	ErrLoginRequired = 1000

	// ErrFullPathOrNameFolderIDNotProvided is returned when no full path or name/folderid provided.
	ErrFullPathOrNameFolderIDNotProvided = 1001

	// ErrFullPathOrFolderIDNotProvided is returned when no full path or folderid provided.
	ErrFullPathOrFolderIDNotProvided = 1002

	// ErrCSROrPublicKeyNotProvided is returned when neither csr or publickey is provided.
	// Please create Certificate Signing Request and pass it as 'csr' parameter or send
	// your 'publickey'.
	ErrCSROrPublicKeyNotProvided = 1003

	// ErrFileIDOrPathNotProvided is returned when no fileid or path provided.
	ErrFileIDOrPathNotProvided = 1004

	// ErrUnknownContentTypeRequested is returned when unknown content-type requested.
	ErrUnknownContentTypeRequested = 1005

	// ErrFlagsNotProvided is returned when please provide flags.
	ErrFlagsNotProvided = 1006

	// ErrInvalidOrClosedFileDescriptor is returned when invalid or closed file descriptor.
	ErrInvalidOrClosedFileDescriptor = 1007

	// ErrLockTypeNotProvided is returned when please provide lock 'type'.
	ErrLockTypeNotProvided = 1008

	// ErrOffsetNotProvided is returned when please provide 'offset'.
	ErrOffsetNotProvided = 1009

	// ErrLengthNotProvided is returned when please provide 'length'.
	ErrLengthNotProvided = 1010

	// ErrCountNotProvided is returned when please provide 'count'.
	ErrCountNotProvided = 1011

	// ErrInvalidLockType is returned when invalid lock type. Please provide type (supported values: 0, 1, 2).
	ErrInvalidLockType = 1012

	// ErrInvalidDateTimeFormat is returned when date/time format not understood.
	ErrInvalidDateTimeFormat = 1013

	// ErrThumbCannotBeCreated is returned when thumb can not be created from this file type.
	ErrThumbCannotBeCreated = 1014

	// ErrInvalidThumbSize is returned when please provide valid thumb size.
	// Width and height must be divisible either by 4 or 5 and must be between 16 and
	// 2048 (1024 for height).
	ErrInvalidThumbSize = 1015

	// ErrFullToPathOrToNameToFolderIDNotProvided is returned when no full topath or toname/tofolderid provided.
	ErrFullToPathOrToNameToFolderIDNotProvided = 1016

	// ErrInvalidFolderID is returned when invalid 'folderid' provided.
	ErrInvalidFolderID = 1017

	// ErrInvalidFileID is returned when invalid 'fileid' provided.
	ErrInvalidFileID = 1018

	// ErrChecksumNotProvided is returned when please provide 'sha1' or 'md5' checksum
	ErrChecksumNotProvided = 1019

	// ErrLanguageNotProvided is returned when please provide language.
	ErrLanguageNotProvided = 1020

	// ErrLanguageNotSupported is returned when language not supported.
	ErrLanguageNotSupported = 1021

	// ErrCodeNotProvided is returned when please provide 'code'.
	ErrCodeNotProvided = 1022

	// ErrMailNotProvidedForShare is returned when please provide 'mail' to share folder with.
	ErrMailNotProvidedForShare = 1023

	// ErrPermissionsNotProvidedForShare is returned when please provide 'permissions' for the share.
	ErrPermissionsNotProvidedForShare = 1024

	// ErrShareRequestIDOrCodeNotProvidedToAcceptShare is returned when please provide 'sharerequestid' or 'code' to accept a share.
	ErrShareRequestIDOrCodeNotProvidedToAcceptShare = 1025

	// ErrShareRequestIDNotProvided is returned when please provide 'sharerequestid'.
	ErrShareRequestIDNotProvided = 1026

	// ErrShareIDNotProvided is returned when please provide 'shareid'.
	ErrShareIDNotProvided = 1027

	// ErrLinkCodeNotProvided is returned when please provide link 'code'.
	ErrLinkCodeNotProvided = 1028

	// ErrFileIDNotProvided is returned when please provide 'fileid'.
	ErrFileIDNotProvided = 1029

	// ErrLinkIDNotProvided is returned when please provide 'linkid'.
	ErrLinkIDNotProvided = 1030

	// ErrOldPasswordNotProvided is returned when please provide 'oldpassword'.
	ErrOldPasswordNotProvided = 1031

	// ErrNewPasswordNotProvided is returned when please provide 'newpassword'.
	ErrNewPasswordNotProvided = 1032

	// ErrMailNotProvided is returned when please provide 'mail'.
	ErrMailNotProvided = 1033

	// ErrPasswordNotProvided is returned when please provide 'password'.
	ErrPasswordNotProvided = 1034

	// ErrCommentNotProvided is returned when please provide 'comment'.
	ErrCommentNotProvided = 1035

	// ErrUploadLinkIDNotProvided is returned when please provide 'uploadlinkid'.
	ErrUploadLinkIDNotProvided = 1036

	// ErrToPathToFolderIDOrToNameNotProvided is returned when please provide at least one of 'topath', 'tofolderid' or 'toname'.
	ErrToPathToFolderIDOrToNameNotProvided = 1037

	// ErrFileIDsNotProvided is returned when please provide 'fileids'.
	ErrFileIDsNotProvided = 1038

	// ErrNameNotProvided is returned when please provide 'name'.
	ErrNameNotProvided = 1039

	// ErrURLNotProvided is returned when please provide 'url'.
	ErrURLNotProvided = 1040

	// ErrMessageNotProvided is returned when please provide 'message'.
	ErrMessageNotProvided = 1041

	// ErrReasonNotProvided is returned when please provide 'reason'.
	ErrReasonNotProvided = 1042

	// ErrUploadNotFound is returned when upload not found
	ErrUploadNotFound = 1900

	// ErrLoginFailed is returned when log in failed.
	ErrLoginFailed = 2000

	// ErrInvalidFileOrFolderName is returned when invalid file/folder name.
	ErrInvalidFileOrFolderName = 2001

	// ErrComponentOfParentDirectoryNotExists is returned when a component of parent directory does not exist.
	ErrComponentOfParentDirectoryNotExists = 2002

	// ErrAccessDenied is returned when access denied. You do not have permissions to preform this operation.
	ErrAccessDenied = 2003

	// ErrFileOrFolderAlreadyExists is returned when file or folder alredy exists.
	ErrFileOrFolderAlreadyExists = 2004

	// ErrDirectoryNotExists is returned when directory does not exist.
	ErrDirectoryNotExists = 2005

	// ErrFolderNotEmpty is returned when folder is not empty.
	ErrFolderNotEmpty = 2006

	// ErrCannotDeleteRootFolder is returned when cannot delete the root folder.
	ErrCannotDeleteRootFolder = 2007

	// ErrUserOverQuota is returned when user is over quota.
	ErrUserOverQuota = 2008

	// ErrFileNotFound is returned when file not found.
	ErrFileNotFound = 2009

	// ErrInvalidPath is returned when invalid path.
	ErrInvalidPath = 2010

	// ErrRequestedSpeedLimitTooLow is returned when requested speed limit too low, see minspeed for minimum.
	ErrRequestedSpeedLimitTooLow = 2011

	// ErrInvalidCodeProvided is returned when invalid 'code' provided.
	// This may happen during two-factor authentication.
	ErrInvalidCodeProvided = 2012

	// ErrEmailAlreadyVerified is returned when email is already verified.
	ErrEmailAlreadyVerified = 2013

	// ErrEmailVerificationRequired is returned when please verify your email address to perform this action.
	ErrEmailVerificationRequired = 2014

	// ErrCannotShareRootFolder is returned when can not share root folder.
	ErrCannotShareRootFolder = 2015

	// ErrCannotShareAlienFolders is returned when you can only share your own folders.
	ErrCannotShareAlienFolders = 2016

	// ErrUserRejectsShares is returned when user does not accept shares.
	ErrUserRejectsShares = 2017

	// ErrInvalidMail is returned when invalid 'mail' provided.
	ErrInvalidMail = 2018

	// ErrShareRequestAlreadyExists is returned when share request already exists.
	ErrShareRequestAlreadyExists = 2019

	// ErrCannotShareWithOneself is returned when you can't share a folder with yourself.
	ErrCannotShareWithOneself = 2020

	// ErrNonExistingShareRequest is returned when non existing share request.
	// It might be already accepted or cancelled by the sending user.
	ErrNonExistingShareRequest = 2021

	// ErrWrongUserForShare is returned when wrong user to accept the share.
	ErrWrongUserForShare = 2022

	// ErrNestedSharedFolder is returned when you are trying to place shared folder into another shared folder.
	ErrNestedSharedFolder = 2023

	// ErrAccessAlreadyGrantedToFolderOrSubfolder is returned when user already has access to this folder or subfolder of this folder.
	ErrAccessAlreadyGrantedToFolderOrSubfolder = 2024

	// ErrInvalidShareID is returned when invalid shareid.
	ErrInvalidShareID = 2025

	// ErrCannotShareAlienFileOrFolder is returned when you can only share your own files or folders.
	// Copy the file to a folder you own if you need to share it.
	ErrCannotShareAlienFileOrFolder = 2026

	// ErrInvalidOrDeletedLink is returned when invalid or already deleted link.
	ErrInvalidOrDeletedLink = 2027

	// ErrActiveSharesOrShareRequestsOnFolder is returned when there are active shares or sharerequests for this folder.
	ErrActiveSharesOrShareRequestsOnFolder = 2028

	// ErrRevisionNotFound is returned when revision with provided 'revisionid' not found.
	ErrRevisionNotFound = 2029

	// ErrNewPasswordIsSame is returned when new password is the same as the old one.
	ErrNewPasswordIsSame = 2030

	// ErrWrongOldPasswordProvided is returned when wrong 'oldpassword' provided.
	ErrWrongOldPasswordProvided = 2031

	// ErrPasswordTooShort is returned when password too short. Minimum length is 6 characters.
	ErrPasswordTooShort = 2032

	// ErrPasswordCannotStartOrEndWithSpace is returned when password can not start or end with space.
	ErrPasswordCannotStartOrEndWithSpace = 2033

	// ErrPasswordTooSimple is returned when password does not contain enough different characters. The minimum is 4.
	ErrPasswordTooSimple = 2034

	// ErrPasswordWithConsecutiveCharacters is returned when password can not contain only consecutive characters.
	ErrPasswordWithConsecutiveCharacters = 2035

	// ErrVerificationCodeExpired is returned when verification 'code' expired. Please request password reset again.
	ErrVerificationCodeExpired = 2036

	// ErrTermsOfServiceNotYetAccepted is returned when you need to accept Terms of Service and all other agreements to register.
	ErrTermsOfServiceNotYetAccepted = 2037

	// ErrEmailAlreadyRegistered is returned when user with this email is already registered.
	ErrEmailAlreadyRegistered = 2038

	// ErrCannotUploadToAlienFolder is returned when you have to own the folder for upload.
	ErrCannotUploadToAlienFolder = 2039

	// ErrUploadLinkIDNotFound is returned when given 'uploadlinkid' not found.
	ErrUploadLinkIDNotFound = 2040

	// ErrConnectionBroken is returned when connection broken.
	ErrConnectionBroken = 2041

	// ErrCannotRenameRootFolder is returned when cannot rename the root folder.
	ErrCannotRenameRootFolder = 2042

	// ErrCannotMoveFolderToSubfolder is returned when cannot move a folder to a subfolder of
	// itself.
	ErrCannotMoveFolderToSubfolder = 2043

	// ErrVideoLinkForNonVideo is returned when video links can only be generated for videos.
	ErrVideoLinkForNonVideo = 2044

	// ErrTFAExpiredToken is returned when the two-factor authentication token used for TFA login
	// has expired.
	ErrTFAExpiredToken = 2064

	// ErrTFARequired is returned when two-factor authentication is required to login.
	ErrTFARequired = 2297

	// ErrSSLError is returned when sSL error occurred. Check sslerror for more information.
	ErrSSLError = 3000

	// ErrUnableToCreateFileThumb is returned when could not create thumb from the given file.
	ErrUnableToCreateFileThumb = 3001

	// ErrConnectionToiTunesFailed is returned when connection to iTunes failed.
	ErrConnectionToiTunesFailed = 3002

	// ErriTunesError is returned when iTunes error.
	ErriTunesError = 3003

	// ErrTooManyLoginsForIP is returned when too many login tries from this IP address.
	ErrTooManyLoginsForIP = 4000

	// ErrInternalError is returned when internal error. Try again later.
	ErrInternalError = 5000

	// ErrInternalUploadError is returned when internal upload error.
	ErrInternalUploadError = 5001

	// ErrInternalErrorNoServerAvailable is returned when internal error, no servers available. Try again later.
	ErrInternalErrorNoServerAvailable = 5002

	// ErrWriteError is returned when write error. Try reopening the file.
	ErrWriteError = 5003

	// ErrReadError is returned when read error. Try reopening the file.
	ErrReadError = 5004

	// ErrNotModified is returned when not modified.
	ErrNotModified = 6000

	// ErrInvalidLinkCode is returned when invalid link 'code'.
	ErrInvalidLinkCode = 7001

	// ErrLinkDeletedByOwner is returned when this link is deleted by the owner.
	ErrLinkDeletedByOwner = 7002

	// ErrLinkDeletedForCopyrightReasons is returned when this link is deleted bacause of copyright complaint.
	ErrLinkDeletedForCopyrightReasons = 7003

	// ErrLinkExpired is returned when this link has expired.
	ErrLinkExpired = 7004

	// ErrLinkOverTrafficLimit is returned when this link has reached its traffic limit.
	ErrLinkOverTrafficLimit = 7005

	// ErrMaximumDownloadReachesFor is returned when this link has reached maximum downloads.
	ErrMaximumDownloadReachesFor = 7006

	// ErrSpaceLimitForLink is returned when this link has reached its space limit.
	ErrSpaceLimitForLink = 7007

	// ErrFileLimitForLink is returned when this link has reached its file limit.
	ErrFileLimitForLink = 7008
)
