//go:build e2e

package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillRecipeMTLS(t *testing.T) {
	env := setupEnv(t)
	const sslID = "skill-mtls-ssl"

	_, _, _ = runA6WithEnv(env, "ssl", "delete", sslID, "--force")
	t.Cleanup(func() { cleanupSSL(t, sslID) })

	sslJSON := `{
		"id": "skill-mtls-ssl",
		"cert": "-----BEGIN CERTIFICATE-----\nMIIDGTCCAgGgAwIBAgIUPULFnqkki/OuJaooDEeg4rnHE9EwDQYJKoZIhvcNAQEL\nBQAwGzEZMBcGA1UEAwwQdGVzdC5leGFtcGxlLmNvbTAgFw0yNjAzMDcwMTEyMTZa\nGA8yMTI2MDIxMTAxMTIxNlowGzEZMBcGA1UEAwwQdGVzdC5leGFtcGxlLmNvbTCC\nASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMWSMWN7aMLtSt4d4pA5q+OZ\nJxXRcwB5MJcmKdE5aV9LHkj6CT8bIWHHLufngD5L39GQORy6rqaujG+oLgzDPKZZ\nb6LZljxSmDFpKnAJaULvKiJgMIJzk5NpTAAG4t5q2/GMFbyYqFj7onfSgZdIWZj1\n5bSN6P9dJFrM3U2fVtkgX8i55LrE0FGjhS/tE3WzoT8SYU5vGEV50cJkgjARjlKq\nVkinOsLAVMi2dDtIrFRqToYBlv4yxMVDwLBNf7oI+gs1We+MDPDf9bg0cT8VybBW\nzJIb0jTacxUilylAm10apdaWlpVS93b8VNedl+oUxnNRlFUB9ZDsMATJ1Mqbgy0C\nAwEAAaNTMFEwHQYDVR0OBBYEFBryYXCOKLquiNnUYd0k9qSRYYTOMB8GA1UdIwQY\nMBaAFBryYXCOKLquiNnUYd0k9qSRYYTOMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZI\nhvcNAQELBQADggEBAMPcmdJD+icNxBZwPSZDtDwHk4WwXhtTDRlnEHTubhbF/T4k\nlUWU/sLbs8tfNA4zwQewo++jLlVdOKu6fa6lWWfl/0GuSFQe33pRYDg1PXoTXc3v\nUCv2NzkWbu4RFrwclBRyGy5HAch1kmMQf/v2su9WUcq6ncqiwm8eReR9meKa16++\nlWwEoM0RAyLChmthBHHbR+E8eP7EQUj1S7fxB77ivmp4OROoGJINPXKnHJ+9JQKz\n71CzwxA8/Zv591h+yYHACO64kRCguYCugfzG3OWrwyzo5k/+A/dQ6tTvXAgb4NLg\noFLe6ZikPa/hlfEhCoRxDAIWggi1vgmz5NPWfrY=\n-----END CERTIFICATE-----",
		"key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDFkjFje2jC7Ure\nHeKQOavjmScV0XMAeTCXJinROWlfSx5I+gk/GyFhxy7n54A+S9/RkDkcuq6mroxv\nqC4MwzymWW+i2ZY8UpgxaSpwCWlC7yoiYDCCc5OTaUwABuLeatvxjBW8mKhY+6J3\n0oGXSFmY9eW0jej/XSRazN1Nn1bZIF/IueS6xNBRo4Uv7RN1s6E/EmFObxhFedHC\nZIIwEY5SqlZIpzrCwFTItnQ7SKxUak6GAZb+MsTFQ8CwTX+6CPoLNVnvjAzw3/W4\nNHE/FcmwVsySG9I02nMVIpcpQJtdGqXWlpaVUvd2/FTXnZfqFMZzUZRVAfWQ7DAE\nydTKm4MtAgMBAAECggEAN7a3KLeXXkiuMR66UjOBLmh05ikvRfXm5ujDKBYJie7T\n9n5T9zC+ZCVYK9tThb29uwnxoRFfyO81+RKzEbRIuRxFZ+X2AMLF2kEuz3NH9LEe\n75yycUcIWax62xMWDOSTa3U2d/2Qk686LJH3K2MiqQjGovjVuJVaeCSHT6lcQ1Pn\n5iPkcSZIBvMQkNG/CuI2bb/+j6himhFbUTebJ8ln1S6uCpLkBVtj1zI5+SE6j+qk\nZdWVj3uDw7N3gbl8iARSMZyjKvSeafKdzA2chOrs1MFzKlw+ToRmHAGpHgeRBypE\ntgM/wPqYV9bk2Tvlx8aC7FqavsxKdNNZ6EVp5N2F/QKBgQDobh6jhGhXFpu3rSOJ\nCcbbaWDPyL4wk4dSk8+KS2w+PrrJ+ZdRQK8XW4s+dOd67uueqINIomHWIDESu/Jj\n/TliPhJ8KZD/P7lwIAWX/w9ybRKWGpXoYe4ACjDB27pQl7A+FO6H5M0SPLM5dndS\nwO7dT9FkDqpo5t9XhFLfpaEu1wKBgQDZmyHcbJUAWrHtO9UqHQGTAxS7JwpmT4BP\nFGOC5Kbel/dtQ0nxrWeeYf0/HsDGxm3E7XcOVV+3QuLmnuOFGmzaQCb4NW/1tMsL\n3fpXZL2NP6k5HYG4hxn8fb1qUFO7Mr7yN8J9h9WtySN9M3R+GOVEbPb7hUHvzUdV\nNA1f97sxmwKBgBsvLfQv+0gcQ9AituJDO3fUBlenAd+Kkawtz3s8QQeyrIQM5g7B\nwvi3YzzFzYEKSpJ+4QPwwgKaN0MaqknZhwmfeuf8sJG58UVU6XKSiUr3yNG3gEry\nkTR9/J/fxBXC+AD6z78jGn0Ejm2tFl2eZRGLUVEjifjE7+A7gLnZlFV3AoGAAewn\n8W2YJ2eluMXVjUiyUd0uGrUul1bOeGRiuK5Sdxb6naGBjrwMdU7CUQNxipAIOjwq\n2BqS/Oh/XrA6rFteaNM2RO0b7xzIynMMmicOsafFU/bZxYqUBTILMVxCUR4Sp8ss\nUbWYgq+LO7jvp4mKxP79c51qxraWkb8i+x0SL08CgYBc4p3vhGv6N5QDxVI6FLhg\ndwU6UbLSEVLf7WnQ+OHNUH7bCi3ROm/ScAjNatUCdm1r+EK6q0gaXlrtwsSZarXb\n/MxJtMgDTuk9JfQyv/tVHDlrt2fBsJq0tJg0HbguTAbMgAGzxLEkwNrduYBeVPTD\n+QiMywtwPlE5Jmd1d2ZJIQ==\n-----END PRIVATE KEY-----",
		"snis": ["skill-mtls.example.com"]
	}`
	f := writeJSON(t, "ssl", sslJSON)
	stdout, stderr, err := runA6WithEnv(env, "ssl", "create", "-f", f)
	require.NoError(t, err, "ssl create: stdout=%s stderr=%s", stdout, stderr)

	stdout, _, err = runA6WithEnv(env, "ssl", "get", sslID, "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"skill-mtls.example.com"`)
}
