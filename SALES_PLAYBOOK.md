# TokMan Sales Playbook & Enterprise GTM

## Sales Strategy Overview

**Goal**: Build a scalable, repeatable sales process that moves from bottom-up (free tier) to enterprise (high-touch) as market develops.

**Timeline**:
- **Weeks 1-4**: Organic/community (no sales team)
- **Month 2**: SMB sales (1 person, outbound + inbound)
- **Month 3+**: Enterprise GTM (dedicated sales team)

---

## Sales Model

### Channel Mix (Year 1 Revenue)

```
BOTTOM-UP (60% of new customers)
├── Free Tier → Pro Conversion (40%)
│   └── Self-serve, product-driven
├── Marketplace (20%)
│   └── VSCode, GitHub Actions reach
└── Community (0%)
    └── Organic referrals, word-of-mouth

ENTERPRISE (40% of new customers)
├── Outbound Sales (25%)
│   └── Target high-value companies
├── Partnerships (10%)
│   └── IDE/CI-CD platform integrations
└── Inbound (5%)
    └── Referrals, press mentions
```

### Sales Roles Ramp

| Period | SDR | AE | Sales Manager | Notes |
|--------|-----|-----|---------------|-------|
| Month 1 | — | — | — | Founder does all |
| Month 2 | 0.5 | — | — | Part-time outreach |
| Month 3 | 1 | 1 | — | Full team forming |
| Month 4-6 | 2 | 2 | 1 | Full enterprise team |
| Month 9+ | 3 | 3 | 1 | Scaling |

---

## Target Customer Profiles (ICPs)

### SMB / Pro Tier (Primary: Months 2-6)

#### Profile: Engineering Team at Growth Stage SaaS
```
Company Size: 20-200 engineers
Characteristics:
- Heavy Claude/GPT-4 users
- Growing API costs ($500-2,000/month)
- Technical decision makers (eng leads)
- Want to save money immediately

Buying Signal:
- Mention of API costs in hiring posts
- Recently adopted Claude
- Growth phase (hiring engineers)
- Using GitHub Actions

Pain Point:
- Monthly API bill growing (>10% month-over-month)
- Budget pressure from CFO
- Need easy solution (no engineering work)

Decision Maker: Engineering Lead, Tech Lead
Deal Size: $99/month = $1,188/year
Sales Cycle: Self-serve (1-2 days to decision)
```

#### Profile: Data Science / ML Teams
```
Company Size: Any
Characteristics:
- High token usage (ML training logs, data analysis)
- Analytical mindset
- Budget conscious
- Early adopters

Buying Signal:
- Large notebook files
- Data science workflows
- Python/R heavy codebases

Pain Point:
- Expensive context windows for model training
- Data preprocessing costs
- Model evaluation costs

Decision Maker: ML Team Lead, Analytics Manager
Deal Size: $99-299/month
Sales Cycle: 2-4 weeks (team discussion)
```

### Enterprise (Primary: Month 6+)

#### Profile: Large Tech Company (1000+ engineers)
```
Company Size: 1,000+ engineers
Characteristics:
- Significant API spend ($10K-100K+/month)
- Procurement process
- Security/compliance focused
- Strategic partnerships value

Buying Signal:
- Large teams using Claude/GPT-4
- Request for on-premise deployment
- NDA/security questions early
- Want dedicated support

Pain Point:
- Governance of AI API usage
- Cost control across teams
- Compliance and security requirements
- Need for usage analytics

Decision Maker: VP Engineering, CTO
Additional Stakeholders: Security, Procurement, Finance
Deal Size: $10K-50K/year
Sales Cycle: 3-6 months
Implementation: 2-4 weeks
```

#### Profile: AI-First Company
```
Company Size: 50-500 engineers
Characteristics:
- Core business depends on AI/Claude
- Product integrates Claude API
- Fast decision making
- High API spend

Buying Signal:
- Claude integrated in product
- Rapid growth
- Cost optimization important for margins
- Willing to customize

Pain Point:
- API costs critical to unit economics
- Need tighter token usage
- Want custom filters for domain
- Competitive advantage through efficiency

Decision Maker: VP Product, VP Engineering
Deal Size: $5K-30K/year (or usage-based)
Sales Cycle: 4-8 weeks
Implementation: 1-2 weeks
```

---

## Sales Messaging & Positioning

### Value Prop by Customer Type

#### For Engineers (SMB, Pro Tier)
```
"Cut your Claude API costs by 75% in 2 minutes"

Headline: Painless cost reduction that just works
Value: Immediate savings on monthly bill
Proof: Real results from 100 beta users
CTA: Install from VSCode Marketplace (free trial)

Positioning: Magic button for AI costs
Tone: Quick, easy, technical
```

#### For Engineering Leaders (Mid-Market)
```
"Reduce AI API costs across your entire team"

Headline: Control team spending without sacrificing productivity
Value: 
- 60-90% cost reduction
- Team-wide visibility
- Automatic budget enforcement
Proof: Used by 50+ companies, $150K+ in savings
CTA: Schedule demo with pricing

Positioning: Cost governance tool for AI teams
Tone: Professional, data-driven, strategic
```

#### For Enterprise (Compliance, Security)
```
"Enterprise-grade token optimization with full control"

Headline: Cost optimization that meets your security & compliance needs
Value:
- On-premise deployment
- HIPAA/SOC2 ready
- Custom compliance
- Dedicated support
Proof: Trusted by [company], [company]
CTA: Discussion with enterprise team

Positioning: Strategic tool for AI cost management
Tone: Trusted advisor, solution-oriented
```

---

## Sales Pitch Template (5 Minutes)

### Opening (30 seconds)
```
"Thanks for taking the time to chat. I know everyone's feeling the pressure 
around AI API costs—whether it's Claude, GPT-4, or others.

We built TokMan to solve exactly that. In the last 8 weeks, 100 companies 
using TokMan cut their API costs by an average of 76%.

I wanted to talk to you because I think there's a real opportunity here 
for [Your Company]."
```

### Problem Statement (1 minute)
```
"From what I understand about [Your Company], you're:
1. Using Claude/GPT-4 heavily (correct?)
2. Seeing API costs grow month over month
3. Want to optimize without hiring more engineers

Is that accurate? What's your biggest concern around API costs right now?"

[Listen. Understand their specific pain.]
```

### Solution (2 minutes)
```
"What TokMan does is apply 30+ compression techniques to your code context. 
Think of it as removing the noise from what you send to Claude.

For example:
- Remove verbose error messages (keep just the key error)
- Strip unnecessary comments (we know it's a function)
- Compress redundant type definitions
- Optimize imports

The result? You send the same information, Claude understands it just as well, 
but you use 60-90% fewer tokens.

It works as a:
1. VSCode extension (real-time preview)
2. GitHub Action (in CI/CD)
3. Programmatic API (if you need it)

So it's frictionless for your team."
```

### Proof (1 minute)
```
"We spent 8 weeks validating this with 100 beta users across different 
use cases—data science teams, backend engineers, API development.

Results:
- 76% average compression ratio
- $150K+ in estimated savings across all users
- 4.7 out of 5 satisfaction
- 92% said they'd recommend it

One customer told us: 'TokMan cut our API costs by 75% immediately. 
We're saving $400/month and the adaptive learning keeps improving week over week.'"
```

### Close (30 seconds)
```
"Here's what I'm thinking—no pressure, but I'd love to:
1. Show you a quick 10-minute demo in your VSCode
2. Have you analyze 1-2 files from your codebase
3. See actual savings numbers for your code

Would next Tuesday or Wednesday work for a 15-minute call to dive deeper?"
```

---

## Cold Outreach Email (SMB)

```
Subject: Cut your Claude costs by 75% (one example)

Hi [First Name],

I saw you're using Claude heavily at [Company] (great choice!).

We just launched TokMan—a tool that reduces token usage 60-90% by compressing 
code context automatically.

Real example from a beta user:
- Before: 12,500 tokens for code analysis
- After: 2,250 tokens (same analysis)
- Savings: $0.31 per request, $100+/month

How it works:
1. Install VSCode extension (2 minutes)
2. Hover over code (see token savings)
3. Use compressed code with Claude
4. Repeat

Takes zero engineering work, just install and save.

We validated this with 100 companies last month—average 76% compression.

Want to give it a try? It's free for the first month:

[Install from Marketplace]

Or if you want to chat first:
[Schedule 15-min call]

No pressure either way. Either way, I think you'll find the compression 
quality impressive.

Cheers,
[Your Name]
TokMan
```

---

## Cold Outreach Email (Enterprise)

```
Subject: Enterprise AI cost optimization—quick question

Hi [First Name],

I'm reaching out because TokMan is launching to the market next week, and 
based on what I know about [Company], I think there's a real opportunity for you.

Quick context: We help companies reduce AI API costs by 60-90% through 
intelligent code compression. Over 8 weeks, we validated the product with 
100 customers—average 76% compression, 92% would recommend.

For a company like [Company] with [estimated high API spend], even a 60% 
reduction could save $50K+ annually.

Here's what makes TokMan different:
✓ Code-specific (not generic prompt optimization)
✓ IDE integration (VSCode, JetBrains coming)
✓ Enterprise-ready (on-premise, HIPAA, audit logs)
✓ Adaptive learning (improves per codebase)
✓ Works with any LLM (Claude, GPT-4, etc.)

I'd love to have a brief conversation about:
1. Your current Claude/GPT-4 usage and spend
2. Cost optimization priorities
3. Fit for TokMan at [Company]

Would you have 20 minutes next week? I'm flexible on timing.

Best,
[Your Name]
TokMan
[Phone] | [Email]

P.S. If enterprise features are important to you, we offer on-premise 
deployment, dedicated support, and custom SLAs.
```

---

## Sales Discovery Call Script

### Opening (2 minutes)
```
"Thanks for taking the time. Before we dive in, I want to understand 
your world first—what's driving your interest in TokMan?"

[Listen to their answer. It reveals priorities.]

"Got it. And when you say [pain point], what does that look like 
concretely? Like, monthly costs, percentage of budget, etc?"

[Get specific numbers. Paint the picture with them.]
```

### Discovery Questions (8 minutes)
```
1. "How are you currently using Claude/GPT-4?"
   └─ Get: Use case, frequency, volume

2. "What's your API spend looking like currently?"
   └─ Get: Monthly costs, growth trend

3. "How much of that is essential vs. could be optimized?"
   └─ Get: Their belief about waste factor

4. "Who else needs to be involved in a decision?"
   └─ Get: Decision committee

5. "What would need to be true for this to be a no-brainer?"
   └─ Get: Criteria for success

6. "What's the biggest concern you'd have about implementing 
   something new?"
   └─ Get: Risk factors to address
```

### Demo (3-5 minutes)
```
"What I'd like to do is show you TokMan in action using your code. 
Can you paste a function or file you're working with?"

[Analyze with TokMan, show:
1. Original token count
2. After compression
3. Compression ratio %
4. Potential monthly savings

This is powerful because it's THEIR code, THEIR savings.]

"What do you think? Does that match what you'd hoped to see?"
```

### Objection Handling (2 minutes)

#### "This is a Chrome extension version problem"
```
Response:
"Actually, TokMan is a full IDE integration—VSCode, JetBrains, coming soon. 
Not a chrome extension. It works locally on your machine, and we never 
store your code.

For CI/CD, we have a GitHub Action that runs on your infra, not ours."
```

#### "We need on-premise deployment"
```
Response:
"Great, we offer that for Enterprise customers. Our SaaS is the standard, 
but we have docker-based on-prem for security/compliance requirements.

That's typically part of our Enterprise tier. Let's talk about what you need 
and we can figure out the right approach."
```

#### "We want to do more manual optimization first"
```
Response:
"That makes sense—and you absolutely could. But here's the thing: manual 
optimization is hard to scale across teams and takes engineering time.

TokMan is more like... why not do both? Get the low-hanging fruit automatically 
with TokMan, then your team can focus on strategic optimizations.

What if we just tried it for one month? See what TokMan finds, then decide 
if you want to go deeper?"
```

#### "This seems like it will break things"
```
Response:
"Actually, that's one of the things our beta users were most surprised about. 
The compression is lossless from Claude's perspective—it contains all the 
information, just more efficiently.

We had 100 beta users test this rigorously. Zero data loss, zero broken workflows.

Want me to show you how? Can run it on your actual code, see what it looks like 
before and after?"
```

### Close (1 minute)
```
"Okay, so here's where I'm at: I think TokMan could save you [X]% on costs, 
with zero engineering work on your end.

Next step would be:
1. Try for free this month (no payment info needed)
2. See actual savings on your codebase
3. If it makes sense, move to paid tier

Does that sound reasonable?"

[If yes] "Great! I'll send you the link and some onboarding docs. 
I'll check in with you in a week to see how it's going, okay?"

[If hesitant] "What would help you feel more confident? Any other info 
you'd like?"
```

---

## Inbound Sales Process

### Inbound Lead → Customer

#### Stage 1: Website Inquiry (Same Day)
```
Trigger: Someone fills out "Contact Sales" form

Response:
- Auto-email: Thankyou + link to free trial
- Manual: Sales rep calls within 4 hours (if enterprise-looking)
- Qualification: SMB (self-serve) vs. Enterprise (sales call)

Goal: Get them using product immediately
```

#### Stage 2: Trial Activation (Day 1-3)
```
If they install extension:
- Monitor activation (user logs in)
- 24-hour email: "Getting started" guide
- 48-hour email: "Top tips" for best compression
- 72-hour email: Feedback request ("How's it going?")

If they sign up but don't install:
- 24-hour email: Installation guide (3-step)
- 48-hour email: Video walk-through
- 72-hour email: Offer phone support
```

#### Stage 3: Engagement (Week 1-2)
```
Track:
- Files analyzed (goal: >5)
- Average compression ratio (goal: >60%)
- Dashboard visits (goal: >3)

If low engagement:
- Call to troubleshoot
- Adjust settings
- Show ROI calculator

If high engagement:
- Introduce Pro tier benefits
- Suggest team setup
- Get feedback for case study
```

#### Stage 4: Conversion (Week 2-4)
```
Free → Pro trigger:
- 10+ files analyzed
- Good compression results (>60%)
- Team expressed interest

Conversion email:
Subject: "Your TokMan impact: [X] tokens saved"

Hi [Name],

You've been using TokMan for [X] days and analyzing [X] files. 
Here's your impact:

Tokens Analyzed: 2.5M
Tokens Saved: 1.9M (76%)
Estimated Monthly Savings: $285

Per our calculations, the Pro tier pays for itself in less than 
a day for your usage level.

Ready to unlock unlimited?
[Upgrade to Pro - $99/month]

First month 50% off with code: [CONVERT50]

Questions? [Schedule call]

Cheers,
TokMan Team
```
```

#### Stage 5: Retention (After Conversion)
```
Day 1 after conversion:
- Welcome email for paid tier
- Pro tier feature guide
- Admin setup (team, billing)

Week 1:
- Check in call (especially for early Enterprise users)
- Answer questions
- Gather feedback

Month 1:
- Review usage metrics
- Calculate actual ROI
- Suggest expansion (add seats, upgrades)

Ongoing:
- Monthly business reviews (Enterprise)
- Quarterly check-ins (Pro)
- Community updates
```

---

## Enterprise Sales Process

### Enterprise Lead → Customer (3-6 Months)

#### Stage 1: Qualification (Week 1)
```
Questions to answer:
1. Is this a strategic fit? (Large API spend, governance need)
2. Are the stakeholders aligned? (Eng + Security + Finance)
3. Is there a timeline? (Q1/Q2/Q3 budget cycle)
4. What's the budget? (Annual vs. usage-based)

Red flags (Skip if present):
- "We're just exploring" (no urgency)
- Only one stakeholder interested
- No budget allocated
- Company shrinking, API usage declining

Green flags (Prioritize):
- Immediate pain (API costs out of control)
- Multi-stakeholder interest
- Budget allocated
- Clear ROI calculations
```

#### Stage 2: Discovery & Demos (Week 1-2)
```
Goal: Understand their world deeply

Discovery meeting agenda:
1. Their current state
   - Current API usage and costs
   - Governance challenges
   - Performance targets
   
2. Their desired future
   - Cost reduction targets
   - Team adoption goals
   - Compliance/security needs
   
3. Our capability
   - Live demo on their code
   - ROI calculation
   - Deployment options
   - Support model
   
4. Next steps
   - Pilot program?
   - PoC timeline?
   - Key stakeholders to involve?
```

#### Stage 3: Pilot / PoC (Week 2-6)
```
Scope:
- Free/discounted trial (2-4 weeks)
- Limited team (1 team of 5-10)
- Specific metrics to track
- Weekly check-ins

Success criteria:
- 60%+ compression achieved
- 80%+ team adoption rate
- ROI modeling confirms value
- Zero security/compliance issues

Outcome:
- Go/no-go decision
- Path to expansion (more teams)
- Commercial terms negotiation
```

#### Stage 4: Commercial Negotiation (Week 6-8)
```
Pricing models:
1. Seat-based ($99/user/month) - most common
2. Usage-based ($0.10-0.30 per 1M tokens saved) - for high-volume
3. Flat annual ($25-50K) - for enterprise simplicity
4. Hybrid (base + usage) - for transparency

Typical enterprise deal:
- $5K-30K annually
- 1-3 year contract
- Renewal escalation clause
- Volume discounts
- Custom SLA

Terms to negotiate:
- Number of users
- Deployment (cloud vs. on-prem)
- Support SLA (4-hour, 8-hour, etc.)
- Custom features (rare, but possible)
- Renewal terms
```

#### Stage 5: Implementation (Week 8-12)
```
Timeline:
- Week 1: Kickoff meeting, user setup
- Week 2-3: Team training
- Week 4: Go-live
- Week 5+: Monitoring, optimization

Deliverables:
- ✓ Deployment (cloud or on-prem)
- ✓ User provisioning
- ✓ Security/compliance documentation
- ✓ Training (recorded + live session)
- ✓ 30-day success check-in

Success metrics:
- 90%+ team adoption
- 60%+ compression (or better)
- Zero production issues
- Customer satisfaction (NPS 50+)
```

#### Stage 6: Account Management (Ongoing)
```
QBR (Quarterly Business Review):
- Metric review (tokens, cost savings, adoption)
- Usage trends analysis
- Roadmap updates
- ROI recalculation
- Expansion opportunities (new teams)

Expansion motions:
- Add more teams
- Expand to other divisions
- Increase usage limits
- Enterprise support upgrade
- Custom feature requests
```

---

## Sales Compensation Plan (Year 1)

### Sales Role Structure
```
SDR (Sales Development Rep):
- Target: 5 qualified meetings/month per SDR
- Compensation: $40K base + $500 per qualified meeting
- Quota: 60 qualified meetings/year

Account Executive (AE):
- Target: 2 Pro customers/month, 1 Enterprise customer/quarter
- Compensation: $80K base + 10% commission on ACV
  - Pro: $99 × 12 × 10% = $119 per close
  - Enterprise: $15K ACV × 10% = $1,500 per close
- Quota: $500K ACV for year

Sales Manager:
- Target: Manage 3-4 AEs, $2M+ ACV total
- Compensation: $100K base + 5% commission on team ACV
- Bonus: Hit quota → +$20K
```

### Commission Examples

#### Pro Customer (SMB)
```
Deal: $99/month × 12 months = $1,188 ACV
Commission: $1,188 × 10% = $119
Annual impact: Modest for AE, huge in volume
```

#### Enterprise Customer
```
Deal: $20,000/year
Commission: $20,000 × 10% = $2,000
Lifetime (assume 3-year retention): $6,000
Impact: Significant, motivates enterprise focus
```

---

## Sales Ops & Infrastructure

### CRM Setup
- **Platform**: Salesforce or HubSpot
- **Key Objects**: Leads, Opportunities, Accounts, Contacts
- **Automation**: Email sequences, task creation, pipeline management
- **Reporting**: Win rate, cycle time, CAC, LTV

### Sales Cadence
- **Weekly**: Pipeline review (deals, movement)
- **Monthly**: Sales meeting (wins, misses, coaching)
- **Quarterly**: Planning and forecasting

### Metrics to Track
```
Input Metrics (leading indicators):
- Outreach (emails, calls, meetings)
- Conversation quality (discovery, demo)
- Proposal rate (% of discovery → proposal)

Output Metrics (trailing indicators):
- Pipeline creation ($M in open opportunities)
- Win rate (% of proposals → closed)
- Sales cycle (days from lead to close)
- CAC (cost per customer acquired)
- LTV:CAC ratio (target: >3:1, aim for 10:1+)
```

---

## Sales Team Hiring Timeline

### Month 2: First Sales Hire
```
Role: Account Executive (0.5 FTE → 1 FTE)
Requirements:
- 2+ years SaaS sales experience
- Developer tool knowledge (bonus)
- Self-motivated, entrepreneurial
- Comfortable with unpredictable product

Responsibilities:
- Pro tier sales (SMB outbound + inbound)
- Enterprise discovery meetings
- Customer success follow-up

Hiring timeline: 2-3 weeks
Start date: Week 6-8
Ramp time: 4-6 weeks to productivity
```

### Month 3: Second Sales Hire
```
Role: Sales Development Rep (SDR)
Requirements:
- 1+ years sales or customer success experience
- Outbound experience (email, calls)
- Persistence and organization
- Growth mindset

Responsibilities:
- Inbound lead qualification
- Outbound prospecting
- Discovery meetings (qualify)
- Pipeline generation for AEs

Hiring timeline: 2 weeks
Start date: Week 12
Ramp time: 2-3 weeks
```

### Month 6: Sales Manager
```
Role: Sales Manager / Sales Operations
Requirements:
- 3+ years sales experience
- Leadership experience (managed team)
- Strategic thinking
- Revenue responsibility

Responsibilities:
- Manage AEs and SDRs
- Sales strategy and planning
- Pipeline and forecasting
- Coaching and development

Hiring timeline: 4 weeks
Start date: Month 6
Focus: Build sustainable sales organization
```

---

## Sales Success KPIs (Year 1)

| Metric | Q1 | Q2 | Q3 | Q4 |
|--------|-----|-----|-----|------|
| **Pipeline Created** | $100K | $300K | $600K | $1M+ |
| **Wins (All Tiers)** | 2 | 15 | 40 | 100+ |
| **Pro Customers** | 0 | 10 | 30 | 50+ |
| **Enterprise Pilots** | 0 | 2 | 4 | 5+ |
| **ARR** | $0 | $15K | $40K | $500K+ |
| **Win Rate** | N/A | 20% | 25% | 30% |
| **Sales Cycle** | — | 30d | 25d | 20d |
| **CAC** | N/A | $150 | $100 | $50 |

---

**TokMan Sales Playbook: From Founder Sales to Scalable Organization** 📈

*Turn product-market fit into revenue growth with a repeatable sales machine.*
