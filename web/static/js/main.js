function callEndpoint(methode, url, body = null) {
  const options = {
    method: methode,
    mode: 'same-origin',
    cache: 'no-cache',
    credentials: 'same-origin',
    redirect: 'follow',
    referrerPolicy: 'no-referrer',
    headers: {
      Authorization: `Bearer ${BearerToken}`,
    }
  }

  if (body !== null) {
    options.headers['accept'] = 'application/json';
    options.headers['content-type'] = 'application/json';
    options.body = JSON.stringify(body)
  };
  return fetch(url, options);
}

function submitVote(vote) {
  callEndpoint('PUT', `/session/${SessionID}/${vote}`);
  getStatus();
}

function resetSession() {
  callEndpoint('GET', `/session/${SessionID}/reset`);
  getStatus();
}

function administer() {
  callEndpoint('PUT', `/session/${SessionID}`, administerData);
  administerData.ShouldShow = !administerData.ShouldShow;
  getStatus();
}

function deleteSession() {
  callEndpoint('DELETE', `/session/${SessionID}`);
  window.location.replace("/");
}

function getStatus() {
  callEndpoint('GET', `/session/${SessionID}/status`).then(res => {
    if (res.ok) {
      return res.json()
    } else {
      throw Error(res.statusText);
    }
  }).then(
    data => {
      voteResultsDiv.innerHTML = '';
      var child;
      data.Votes.forEach(element => {
        child = document.createElement('div');
        child.innerText = element.Name + ' ' + element.Vote;
        voteResultsDiv.appendChild(child);
      });

      voteResultsDiv.appendChild(document.createElement('br'));

      child = document.createElement('div');
      child.innerText = 'Average ' + data.Average;
      voteResultsDiv.appendChild(child);

      child = document.createElement('div');
      child.innerText = 'Closest ' + data.Closest;
      voteResultsDiv.appendChild(child);
    }
  ).catch(e => { console.log(e) });
}

const voteResultsDiv = document.getElementById('voteResults');
const administerData = {
  'ShouldShow': true
};

const myTimeout = setInterval(getStatus, votePullInterval);
getStatus();