/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type RequestAttributes struct {
	// Module to grant permission
	Module string `json:"module"`
	// Already built payload to grant permission <br><br> -> \"add_user\" = action to add user in repository or group in gitlab<br> -> \"verify_user\" = action to verify user in gitlab module (connect user id from identity with gitlab username)<br> -> \"update_user\" = action to update user access level in repository or group in gitlab<br> -> \"get_users\" = action to get users with their permissions from repository or group in gitlab<br> -> \"delete_user\" = action to delete user from module (from all links)<br> -> \"remove_user\" = action to remove user from repository or group in gitlab<br>
	Payload json.RawMessage `json:"payload"`
}
