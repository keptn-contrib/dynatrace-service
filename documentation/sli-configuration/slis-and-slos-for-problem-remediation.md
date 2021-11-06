## SLIs & SLOs for Problem Remediation

If Dynatrace sends problems to Keptn which triggers an Auto-Remediation workflow, Keptn also evaluates your SLOs after the remediation action was executed.
The default behavior that users expect is that the auto-remediation workflow can stop if the problem has been closed in Dynatrace and that it should continue otherwise!

When a Dynatrace Problem initiates a Keptn auto-remediation workflow the *dynatrace-service* adds the Dynatrace Problem URL as a label with the name "Problem URL". As labels are passed all the way through every event along a Keptn process it also ends up being passed as part of the `sh.keptn.internal.event.get-sli` which is handled by the *dynatrace-service*.

Here is an excerpt of that event showing the label:

```json
 "labels": {
      "Problem URL": "https://abc12345.live.dynatrace.com/#problems/problemdetails;pid=3734886735257827488_1606270560000V2",
      "firstaction": "action.triggered.firstaction.sh"
    },
    "project": "demo-remediation"
```

So, if the *dynatrace-service* detects that it gets called in context of a remediation workflow and finds a Dynatrace Problem ID (PID) as part of the Problem URL it will query the status of that problem (OPEN or CLOSED) using Dynatrace's Problem API v2. It will then return an SLI called `problem_open` and the value  either be `0` (=problem no longer open) or `1` (=problem still open).

The *dynatrace-service* will also define a key SLO for `problem_open` with a default pass criteria of `<=0` meaning the evaluation will only succeed if the problem is closed. The following is an excerpt of that SLO definition:

```yaml
objectives:
- sli: problem_open
  pass:
  - criteria:
    - <=0
  key_sli: true
```

As the SLO gets added if it's not defined and as the sli named `problem_open` will always be returned this capability allows you to either define your own custom SLO including `problem_open` as an SLO or you just go with the default that *dynatrace-service* creates.