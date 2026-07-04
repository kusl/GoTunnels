free tier only 

I have this activity.html 
it has a problem in that it is not very mobile friendly
what is the best way with best in class engineering practices
to make the website responsive across different screen sizes 
and device capabilities without adding a build step 
or adding any vendor software or packages? 
Please give me all the possible options 
along with the proposed code implementations 

```html
<!doctype html>
<html lang="en"><head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Activity · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css">
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/activity" data-auth="in" class="active">Activity</a>
          <a href="/settings" data-auth="in" class="">Settings</a>
          <a href="/login" data-auth="out" class="hidden">Log in</a>
          <a href="/signup" data-auth="out" class="hidden">Sign up</a>
        </nav>
      </div>
    </header>

    <main>
      <h1>Your activity</h1>
      <p class="lead">
        Every sign-up and login is recorded. Your IP address is never stored —
        only a salted SHA-256 hash of it, shown here exactly as stored.
      </p>

      <div id="msg" class="msg"></div>

      <div class="card">
        <table>
          <thead>
            <tr>
              <th>When</th>
              <th>Event</th>
              <th>Method</th>
              <th>Outcome</th>
              <th>IP hash</th>
            </tr>
          </thead>
          <tbody id="activity-body"><tr><td>7/4/2026, 8:31:51 AM</td><td>login</td><td>password</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr><tr><td>7/4/2026, 8:31:48 AM</td><td>logout</td><td>—</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr><tr><td>7/4/2026, 8:31:36 AM</td><td>signup</td><td>password</td><td><span class="tag success">success</span></td><td class="hash">1a96e1c8f68f0edbb602f8adc19b4b715e832504a17e2660a45d3ad0f87875cb</td></tr></tbody>
        </table>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <script type="module" src="/js/page-activity.js"></script>
  

</body></html>
```

Gemini Pro Extended (free tier)
