package build

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/datadog"
	"github.com/Nextdoor/conductor/shared/flags"
	"github.com/Nextdoor/conductor/shared/types"
)

var (
	jenkinsURL      = flags.EnvString("JENKINS_URL", "")
	jenkinsUsername = flags.EnvString("JENKINS_USERNAME", "")
	jenkinsPassword = flags.EnvString("JENKINS_PASSWORD", "")
	jenkinsService  *jenkins
)

func Jenkins() Service {
	if jenkinsService == nil {
		// Initialize Jenkins.
		if jenkinsURL == "" {
			panic(errors.New("jenkins_url flag must be set."))
		}
		if jenkinsUsername == "" {
			panic(errors.New("jenkins_username flag must be set."))
		}
		if jenkinsPassword == "" {
			panic(errors.New("jenkins_password flag must be set."))
		}

		jenkinsService = &jenkins{
			URL:      jenkinsURL,
			Username: jenkinsUsername,
			Password: jenkinsPassword}

		err := jenkinsService.TestAuth()
		if err != nil {
			panic(err)
		}
	}

	return jenkinsService
}

type jenkins struct {
	URL      string
	Username string
	Password string
}

func (j jenkins) TestAuth() error {
	baseUrl, err := url.Parse(j.URL)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)

	resp, err := j.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error connecting to Jenkins: %s", resp.Status)
	}
	return nil
}

func (j jenkins) TriggerJob(job *types.Job, jobName string, params map[string]string) {
	datadog.Info("Triggering Jenkins Job \"%s\", Params: %s", jobName, params)

	// How do we handle these errors? Send a slack message?
	//   What would someone do after this slack message?
	// Auto retry? (probably for some of them)
	// Rebuild button on Conductor UI (not for MVP).
	buildUrl, err := url.Parse(fmt.Sprintf("%s/job/%s/buildWithParameters", j.URL, jobName))
	if err != nil {
		return
	}

	dataClient := data.NewClient()
	err = dataClient.TriggerJob(job)
	if err != nil {
		return
	}

	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v)
	}
	buildUrl.RawQuery = urlParams.Encode()

	req, err := http.NewRequest("POST", buildUrl.String(), nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(j.Username, j.Password)

	resp, err := j.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != 201 {
		return
		// return fmt.Errorf("Error building Jenkins job: %s", resp.Status)
	}
	return
}

func (j jenkins) PollJobById(jobName, buildId string) error {

}

func (j jenkins) PollJobsByName(jobName string) error {
	// Find all running jobs by name.
	// If it's a new job, we check its SHA and find the phase for it.
	// Check if each one is a new job or not (by build id).
	// If it's already being tracked, we just check if it's already done or not.
	// What if the job finishes before ever being polled and going through the phase transition?
	// triggerJob should probably watch it... but then how do we do the error handling?
	// tbh, the error handling is already pretty bad. What happens when TriggerJob fails today?
	// phase.Error... which does nothing.
}

func (j jenkins) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 15,
	}
	return client.Do(req)
}

/*
   async def build_job(self, job_name, params):
       build_url = f'{self.url}/job/{job_name}/buildWithParameters?{urllib.parse.urlencode(params)}'
       status, response, headers = await self.do_request('post', build_url)
       if status != 201:
           raise JenkinsError(f'Could not start Jenkins job: {status}\n\n{response}')

       return headers['Location'].replace('http://', 'https://')

   async def find_job(self, queue_url):
       queue_url += 'api/json'
       while True:
           status, response, _ = await self.do_request('get', queue_url)
           if status != 200:
               continue
           if response is not None and 'executable' in response:
               if 'url' in response['executable']:
                   return response['executable']['url'].replace('http://', 'https://')
           await asyncio.sleep(1)

   async def wait_for_job_completion(self, job_url):
       job_url += 'api/json'
       while True:
           status, response, _ = await self.do_request('get', job_url)
           if status != 200:
               continue
           if response is not None and response['result'] is not None:
               return response['result']
           await asyncio.sleep(1)

   @retry(max_retries=5, retry_delay_secs=2)
   async def do_request(self, method, url):
       async with aiohttp.ClientSession() as session:
           async with getattr(session, method)(url, auth=self.auth, timeout=aiohttp.ClientTimeout(total=20)) as response:
               if response.status == 200:
                   return response.status, await response.json(), response.headers
               elif response.status < 400:
                   return response.status, await response.text(), response.headers
               else:
                   text = await response.text()
                   raise JenkinsError(f'{response.status}: {text}')
*/
