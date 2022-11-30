---
name: Epic template
about: Create a new Epic
labels: ["🌠 epic"]
---
## What is it about?

<!-- Description, motivation or user story -->

## Acceptance criteria

- [ ] xxx
- [ ] yyy
- [ ] zzz

## Dependencies

needs #...
needed by #...
relates to #...

## Risk assessment

<!-- Be specific to this change. General threats are documented in the threat modeling document -->

Describe risks specific to this change, for example:

- Failure modes
- Security risks
- performance impact

Also describe how these risks can be detected through alerting or testing.
Finally describe the mitigations for these risks.

### Risk 1

* ...

Mitigations
* ...


## Verification checklist

The change was verified, tested and deployed to the following stages:

- [ ] Development
- [ ] Integration
- [ ] Production

Security tests:

- [ ] Vulnerability scan
- [ ] Penetration test

_Which security tests are mandetory or optional depends on the complexity label, see the [secure development policy](https://confluence.camunda.com/display/ISMS/P11+-+Secure+Software+Development#P11SecureSoftwareDevelopment-Phase1:ChangeDiscovery) for more information_

## Definition of Ready - Checklist

_Not all items need to be done depending on the issue._

- [ ] the issue has a meaningful title, description and testable acceptance criteria
- [ ] risks and mitigations are identified
- [ ] Added a complexity label following the [secure development policy](https://confluence.camunda.com/display/ISMS/P11+-+Secure+Software+Development#P11SecureSoftwareDevelopment-Phase1:ChangeDiscovery)
- [ ] unclear requirements have been broken down into separate issues (always make progress)
- [ ] discussed with the team
