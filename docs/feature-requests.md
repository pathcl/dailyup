# Feature Request Framework

This document defines how to propose, evaluate, and prioritise new features in
a structured way. The goal is to move decisions away from opinion and toward
evidence, and to ensure every feature can demonstrate value before and after it
ships.

---

## Feature Request Template

Fill in all five sections. If you cannot answer any one of them concretely, the
idea is not ready to evaluate yet — gather more information first.

```markdown
## Feature: <name>

### Problem statement
Who has this problem, how often, and what do they do today instead?
One concrete sentence is better than a paragraph.
Example: "Sales reps spend 20 min/week manually exporting X to Y because Z."

### Value hypothesis
Which of these does it deliver? (check all that apply)
- [ ] Revenue impact — enables sales, reduces churn
- [ ] Cost reduction — automates manual work, reduces support load
- [ ] Risk / compliance — avoids fines, meets a regulatory requirement
- [ ] Strategic — unlocks a market segment, required for a partnership

### Success metric
How do we know it worked? Define this before building, not after.
Example: "Support tickets about X drop 30% within 60 days of release."

### Scope
- Affected users / teams:
- Frequency of use: daily / weekly / per-sprint / one-off
- Effort estimate: S / M / L / XL

### Criteria check — all must pass before proceeding
- [ ] We have evidence the problem is real (user research, ticket data, revenue data)
- [ ] Success can be measured with existing tooling
- [ ] It aligns with this quarter's strategic goals
- [ ] Engineering has capacity within the planning window
```

---

## Understanding Value

"Creating value" maps to three distinct types. A strong feature scores on at
least two of the three. Be explicit about which type you are claiming.

| Type | How to measure | Example |
|------|---------------|---------|
| **Customer value** | User accomplishes their goal faster or with less effort | Removes a manual copy-paste step |
| **Business value** | Moves revenue, reduces cost, or reduces risk | Cuts support tickets by 30% |
| **Technical value** | Reduces future development cost or increases reliability | Replaces three fragile scripts with one tested package |

**Common trap:** features that score high on customer value but low on business
value — users like it, the company does not benefit. Optimise for both.

---

## Prioritisation — RICE Score

Use RICE to rank competing features objectively and remove "loudest voice wins"
from the process.

```
Score = (Reach × Impact × Confidence) / Effort
```

| Factor | Definition | Scale |
|--------|-----------|-------|
| **Reach** | Users affected per week | Raw number |
| **Impact** | How much it moves the needle per user | 3 massive / 2 high / 1 medium / 0.5 low / 0.25 minimal |
| **Confidence** | How sure are you of Reach and Impact? | 100% data-backed / 80% anecdotal / 50% gut feel |
| **Effort** | Total person-weeks to design, build, and ship | Raw number |

Higher score = higher priority. Recalculate after each planning cycle as new
data arrives.

---

## Decision Gate

A feature is approved to enter the backlog only when:

1. All five template sections are filled in
2. All four criteria checkboxes are checked
3. The RICE score has been calculated and compared against current backlog items
4. A DRI (Directly Responsible Individual) is named

A feature is rejected (or returned for more information) if:
- No measurable success metric exists
- Evidence of the problem is missing ("I think users want this" is not evidence)
- It requires new credentials, integrations, or infrastructure not already in place
- It duplicates a capability that already exists or is already planned

---

## What Usually Goes Wrong

| Failure mode | What it looks like | Fix |
|---|---|---|
| No success metric | "We'll know it worked when users are happy" | Define a number before building |
| Skipping evidence | Strong opinion presented as fact | Attach data: tickets, interviews, usage logs |
| Confusing shipping with value | "We shipped it, done" | Measure adoption and outcome 30/60/90 days post-launch |
| Scope creep | Feature grows during build | Lock scope at approval; new requests start a new ticket |
| No exit criteria | Build continues indefinitely | Define "good enough" before starting |

---

## Workflow Summary

```
Idea
 └─ Fill in template
     └─ All sections answerable?
         ├─ No  → gather more information, come back later
         └─ Yes → run criteria check
                   └─ All boxes checked?
                       ├─ No  → reject or return with feedback
                       └─ Yes → calculate RICE score
                                 └─ Add to backlog, assign DRI
                                     └─ Build → ship → measure success metric
```
