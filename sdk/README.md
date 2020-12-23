# SDK

# Status

- Intro
  - ䷢ Global Parameters
  - Authentication
    - ✅ Login
    - ✅ TFA login (two-factor authentication)
    - Additional login journeys
- General
  - getdigest
  - ✅ userinfo
  - supportedlanguages
  - setlanguage
  - feedback
  - currentserver
  - ✅ diff
  - ✅ getfilehistory
  - getip
  - getapiserver
- ✅ Folder
  - ✅ createfolder
  - ✅ createfolderifnotexists
  - ✅ listfolder
  - ✅ renamefolder
  - ✅ deletefolder
  - ✅ deletefolderrecursive
  - ✅ copyfolder
- File
  - ✅ uploadfile
  - uploadprogress
  - downloadfile
  - downloadfileasync
  - ✅ copyfile
  - ✅ checksumfile
  - ✅ deletefile
  - ✅ renamefile
  - stat
- Auth
  - sendverificationemail
  - verifyemail
  - changepassword
  - lostpassword
  - resetpassword
  - register
  - invite
  - userinvites
  - ✅ logout
  - ✅ listtokens
  - deletetoken
  - sendchangemail
  - changemail
  - senddeactivatemail
  - deactivateuser
- Streaming
  - ✅ getfilelink
  - getvideolink
  - getvideolinks
  - getaudiolink
  - gethlslink
  - gettextfile
- Archiving
  - getzip
  - getziplink
  - savezip
  - extractarchive
  - extractarchiveprogress
  - savezipprogress
- Sharing
  - sharefolder
  - listshares
  - sharerequestinfo
  - cancelsharerequest
  - acceptshare
  - declineshare
  - removeshare
  - changeshare
- Public Links
  - getfilepublink
  - getfolderpublink
  - gettreepublink
  - showpublink
  - getpublinkdownload
  - copypubfile
  - listpublinks
  - listplshort
  - deletepublink
  - changepublink
  - getpubthumb
  - getpubthumblink
  - getpubthumbslinks
  - savepubthumb
  - getpubzip
  - getpubziplink
  - savepubzip
  - getpubvideolinks
  - getpubaudiolink
  - getpubtextfile
  - getcollectionpublink
- Thumbnails
  - getthumblink
  - getthumbslinks
  - getthumb
  - savethumb
- Upload Links
  - createuploadlink
  - listuploadlinks
  - deleteuploadlink
  - changeuploadlink
  - showuploadlink
  - uploadtolink
  - uploadlinkprogress
  - copytolink
- Revisions
  - listrevisions
  - revertrevision
- Fileops
  - ✅ file_open
  - ✅ file_write
  - file_pwrite
  - ✅ file_read
  - ✅ file_pread
  - ✅ file_pread_ifmod
  - ✅ file_checksum
  - file_size
  - file_truncate
  - ✅ file_seek
  - ✅ file_close
  - file_lock
- Newsletter
  - newsletter_subscribe
  - newsletter_check
  - newsletter_verifyemail
  - newsletter_unsubscribe
  - newsletter_unsibscribemail
- Trash
  - trash_list
  - trash_restorepath
  - trash_restore
  - trash_clear
- Collection
  - collection_list
  - collection_details
  - collection_create
  - collection_rename
  - collection_delete
  - collection_linkfiles
  - collection_unlinkfiles
  - collection_move
- OAuth 2.0
  - authorize
  - oauth2_token
- Transfer
  - uploadtransfer
  - uploadtransferprogress