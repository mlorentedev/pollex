package main

// Sample represents a benchmark text sample.
type Sample struct {
	Name string
	Text string
}

// Samples contains realistic work email texts at varying lengths,
// intentionally written with non-native English mistakes.
// Used by default benchmark mode (--quality=false) for performance measurement.
var Samples = []Sample{
	{
		Name: "tiny",
		Text: "Can you review the PR when you have time? I think is ready but not sure about the error handling part.",
	},
	{
		Name: "short",
		Text: `Hi team,

I wanted to let you know that the deployment yesterday went smooth. All the services are running fine and we didn't saw any errors in the logs so far. The only thing is that the response time for the search endpoint is a bit higher than what we expected, around 450ms instead of 300ms. I will investigate this today and keep you posted.

Thanks,
Manuel`,
	},
	{
		Name: "medium",
		Text: `Hi Sarah,

Following up on our conversation from the standup today. I've been looking into the authentication issue that several users reported last week. After investigating the logs I found that the problem is related to how we handle token refresh when the session expires while the user is in the middle of filling a long form.

What happens is: the token expires, the refresh endpoint returns a new token, but the original request that triggered the refresh gets lost because we don't retry it. This means the user loses their form data which is very frustrating specially for the onboarding form that has like 15 fields.

I think the best approach would be to implement a request queue that holds pending requests while the token is being refreshed, and then replays them once we got the new token. I've seen this pattern in several OAuth libraries and it works good.

I can have a draft PR ready by Thursday if you agree with this approach. Let me know what you think or if you have a different idea.

Best,
Manuel`,
	},
	{
		Name: "long",
		Text: `Subject: Q4 Infrastructure Migration - Status Update and Decision Needed

Hi everyone,

I want to give you all an update on where we are with the infrastructure migration project and also there is a decision we need to make before end of this week.

Current Status:
We have successfully migrated 3 out of 5 services to the new Kubernetes cluster. The API gateway, the notification service, and the user management service are all running in production on the new infrastructure since last Monday. Performance metrics show that the response times are actually better than before, around 15-20% improvement which is great.

However, we hit a blocker with the remaining two services: the billing service and the analytics pipeline.

The billing service has a dependency on a legacy PostgreSQL database that runs on a very old version (9.6) and uses some extensions that are not compatible with the managed database service we planned to use. We have two options here:

Option A: Upgrade PostgreSQL to version 15 first, then migrate. This would take approximately 2 weeks of additional work and involves some risk because we need to test all the billing queries against the new version. The benefit is that we end up with a modern database that is easier to maintain.

Option B: Keep the legacy database running on a dedicated VM alongside the new Kubernetes cluster. This is faster (can be done in 2-3 days) but means we still have to maintain the old infrastructure for this one service indefinitely. Also the networking between the K8s cluster and the VM adds complexity.

For the analytics pipeline, the issue is different. The pipeline processes around 50GB of data daily and the current implementation uses a lot of local disk storage for intermediate results. Moving this to Kubernetes would require us to set up persistent volumes with high IOPS which increases our cloud costs significantly.

My recommendation is to go with Option A for the billing service (take the time to upgrade PostgreSQL properly) and to postpone the analytics pipeline migration until we can refactor it to use streaming processing instead of batch. This way we don't compromise on quality and we avoid the high storage costs.

I've prepared a detailed comparison document that I will share in the team drive later today. Please review it and let me know your thoughts by Friday so we can finalize the migration plan.

Thanks,
Manuel`,
	},
	{
		Name: "max",
		// Quality note: exceeds 1500 char extension limit, for stress testing only.
		Text: `Subject: Post-Incident Report - Production Outage on January 15th

Dear team and stakeholders,

This email serves as the post-incident report for the production outage that occured on January 15th between 14:30 and 16:45 UTC. I want to provide a comprehensive overview of what happened, what was the impact, what we did to resolve it, and most importantly what we will do to prevent similar incidents in the future.

1. Executive Summary

On January 15th, our main application experienced a complete outage lasting approximately 2 hours and 15 minutes. The root cause was a memory leak in the session management module that was introduced in release v2.8.3, deployed that same morning. The leak caused the application servers to run out of memory progressively, starting with the first server going down at 14:30 and the last one at 14:52. During this period, approximately 12,000 users were affected and could not access the platform. Our monitoring system detected the issue 8 minutes after the first server went down, and the on-call engineer was paged at 14:40.

2. Timeline of Events

- 09:15 UTC: Release v2.8.3 was deployed to production through our standard CI/CD pipeline. The release contained 14 changes including a refactoring of the session management module.
- 09:15-14:30 UTC: The application appeared to function normally. However, memory usage on the application servers was gradually increasing at a rate of approximately 50MB per hour per server. Under normal operation, memory usage is stable around 2GB per server.
- 14:30 UTC: Server app-prod-01 ran out of memory (8GB limit) and was killed by the OOM killer. The load balancer detected the health check failure and removed it from the pool.
- 14:35 UTC: Server app-prod-02 started showing degraded performance due to memory pressure (7.2GB used).
- 14:38 UTC: Our monitoring system triggered an alert for "High Memory Usage" on app-prod-02 and app-prod-03.
- 14:40 UTC: On-call engineer (myself) was paged and started investigating.
- 14:45 UTC: I identified the memory growth pattern and suspected the morning release.
- 14:52 UTC: Server app-prod-02 and app-prod-03 ran out of memory simultaneously. Complete outage began.
- 15:00 UTC: I initiated a rollback to v2.8.2. The rollback process started but took longer than expected because our deployment pipeline requires a full build cycle even for rollbacks.
- 15:25 UTC: Rollback deployment completed. Servers started recovering.
- 15:35 UTC: All servers back online and healthy. Load balancer routing traffic normally.
- 15:35-16:45 UTC: Extended monitoring period. We observed some users experiencing session errors because their sessions created during the faulty release were corrupted. We had to implement an emergency session reset for affected users.
- 16:45 UTC: All systems fully operational. Incident declared resolved.

3. Root Cause Analysis

The memory leak was caused by a subtle bug in the session management refactoring. In the new code, every time a user's session was validated (which happens on every API request), a new goroutine was spawned to update the session's "last accessed" timestamp. However, the goroutine was writing to a channel that nobody was reading from in certain edge cases (when the session was about to expire). This caused the goroutines to pile up indefinitely, each holding references to the session data in memory.

The bug was particularly hard to catch because:
- In development and staging environments, the number of concurrent sessions is much lower (hundreds vs tens of thousands), so the leak was too slow to notice.
- Our load tests focus on request throughput and response times, not on memory stability over long periods.
- The code review didn't catch the issue because the channel usage appeared correct at first glance—the bug only manifested when a specific timing condition was met.

4. Impact Assessment

- Duration: 2 hours 15 minutes (14:30-16:45 UTC)
- Users affected: approximately 12,000 (based on unique sessions during that window)
- Revenue impact: estimated $45,000 in lost transactions (based on average hourly revenue for that time period)
- SLA impact: this incident consumed 80% of our monthly error budget
- Customer support: received 340 tickets related to the outage, all resolved within 24 hours

5. Action Items and Prevention

Based on our analysis, we have identified the following action items to prevent similar incidents:

a) Immediate (this week):
- Add memory usage monitoring with automatic alerting when any server exceeds 70% memory (currently set at 90%). This would have given us almost 3 hours of early warning.
- Implement a "quick rollback" mechanism that can revert to the previous release in under 5 minutes, without requiring a full build cycle.

b) Short-term (this month):
- Add a long-running load test to our CI/CD pipeline that runs for at least 2 hours with realistic session counts before any release is approved for production.
- Implement goroutine leak detection in our monitoring stack using runtime metrics.
- Review all channel usage patterns in the codebase for similar unbuffered channel issues.

c) Medium-term (this quarter):
- Implement canary deployments so that new releases are first exposed to only 5% of traffic. If memory usage or error rates deviate from baseline, the deployment is automatically rolled back.
- Establish a formal code review checklist for concurrency-related changes that includes channel lifecycle verification.

I want to acknowledge that our response time could have been better. The 8-minute delay between the first server going down and the alert firing is too long. Also, our rollback process taking 25 minutes is unacceptable for a critical production system. These are the areas where we will see the most improvement from the action items listed above.

If you have questions or suggestions for additional preventive measures, please don't hesitate to reach out. I've scheduled a blameless post-mortem meeting for this Thursday at 10:00 UTC where we can discuss this in detail.

Thank you for your patience and understanding during this incident.

Best regards,
Manuel
Senior Site Reliability Engineer`,
	},
}

// QualitySamples are designed to test specific error types and correction quality.
// Each sample targets a different class of writing issue.
// Used by --quality mode to compare model output quality.
var QualitySamples = []Sample{
	{
		Name: "subtle",
		// Tests: then/than, should of/have, commiting/committing
		Text: "The data suggests that our approach is less effective then we initially thought. We should of considered alternative methods before commiting to this strategy.",
	},
	{
		Name: "technical",
		// Tests: effecting/affecting, wich/which, technical jargon preservation
		Text: "After investigating the incident, we determined that the root cause was a race condition in the authentication middleware which was effecting all requests that relied on the session cache. The fix involves implementing a mutex lock around the shared state, wich prevents concurrent writes from corrupting the token store.",
	},
	{
		Name: "informal",
		// Tests: run-on sentence, tone elevation, dont/don't, alot/a lot
		Text: "So basically what happened is that the client called us yesterday and they were pretty upset about the delay and they said that if we dont deliver by end of month they will cancel the contract which would be really bad for us because this is one of our biggest accounts and we already spent alot of resources on this project so I think we need to have an emergency meeting to figure out how to speed things up.",
	},
	{
		Name: "academic",
		// Tests: wrong prepositions (in/on, for/to), consistant, reccomend, representitive, comprised mostly by
		Text: "The research team has been working in this problem for several months and have produced some preliminary results that are consistant with our hypothesis. However, the sample size is too small to draw definitive conclusions from. We reccomend expanding the study to include participants from different demographics, as the current sample is comprised mostly by college students who may not be representitive of the general population. Additionally, the methodology needs to be revised for addressing the concerns raised by the peer reviewers.",
	},
	{
		Name: "complex",
		// Tests: verbose phrasing, amongst/among, unprecedentedly, innefficiencies, less then, multi-sentence coherence
		Text: "I would like to take this opportunity to bring to your attention the fact that our department has been experiencing a number of significant challenges over the course of the past several months that have had a negative impact on our ability to deliver projects on time and within budget. Firstly, the turnover rate amongst our senior engineers has been unprecedentedly high, which has resulted in a significant loss of institutional knowledge. Secondly, the tools and infrastructure that we are currently utilizing are outdated and no longer fit for purpose, leading to innefficiencies that could easily be avoided with proper investment. Lastly, the communication between our team and the product management department has been less then ideal, with requirements frequently changing midway through development cycles without adequate notice or justification being provided to the engineering team.",
	},
}
