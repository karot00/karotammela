/**
 * Swamp extension model for checking Vercel production logs and alerting via Resend email.
 *
 * @module
 */

import { z } from "npm:zod@4";

const GlobalArgsSchema = z.object({
  vercelToken: z.string().describe("Vercel Personal Access Token or API Key.").meta({ sensitive: true }),
  projectId: z.string().describe("The Vercel Project ID or slug name."),
  resendApiKey: z.string().describe("Resend API Key for sending emails.").meta({ sensitive: true }),
  notificationEmail: z.string().email().describe("The recipient email address for alerts."),
  teamId: z.string().optional().describe("Optional Vercel Team ID or slug if the project belongs to a team."),
});

const LogEventSchema = z.object({
  id: z.string(),
  text: z.string(),
  timestamp: z.number(),
  type: z.string(),
});

const OutputSchema = z.object({
  checkedAt: z.string(),
  totalLogsAnalyzed: z.number(),
  errorsFound: z.number(),
  errorLogs: z.array(LogEventSchema),
});

/** Model definition for Vercel Log Monitoring. */
export const model = {
  type: "@local/vercel-monitor",
  version: "2026.06.02.2",
  globalArguments: GlobalArgsSchema,
  resources: {
    "result": {
      description: "Vercel log analysis and Resend notification state",
      schema: OutputSchema,
      lifetime: "infinite",
      garbageCollection: 10,
    },
  },
  methods: {
    check: {
      description: "Inspect Vercel logs and send email alerts if errors are found.",
      arguments: z.object({}),
      execute: async (args: unknown, context: any): Promise<{ dataHandles: any[] }> => {
        const { vercelToken, projectId, resendApiKey, notificationEmail, teamId } = context.globalArgs;

        const teamParam = teamId ? `&teamId=${teamId}` : "";

        // Step 1: List deployments of the project to find the latest active deployment ID
        const deployUrl = `https://api.vercel.com/v6/deployments?projectId=${projectId}&limit=5${teamParam}`;
        
        context.logger.info("Fetching deployments for project {projectId}...", { projectId });
        const deployRes = await fetch(deployUrl, {
          headers: {
            Authorization: `Bearer ${vercelToken}`,
          },
        });

        if (!deployRes.ok) {
          const errText = await deployRes.text();
          throw new Error(`Failed to list Vercel deployments: status ${deployRes.status} - ${errText}`);
        }

        const deployData = await deployRes.json();
        const deployments = deployData.deployments || [];
        if (deployments.length === 0) {
          throw new Error(`No deployments found for project: ${projectId}`);
        }

        context.logger.info("Found deployments: {deployments}", { 
          deployments: JSON.stringify(deployments.map((d: any) => ({id: d.uid || d.id, state: d.state})))
        });
        
        // Find the latest deployment. By default, Vercel returns them sorted by creation time descending.
        // We'll prioritize READY state if available.
        const latestDeployment = deployments.find((d: any) => d.state === "READY") || deployments[0];
        const deploymentId = latestDeployment.uid || latestDeployment.id;
        
        context.logger.info("Latest deployment is {deploymentId} ({url}) with state {state}", {
          deploymentId,
          url: latestDeployment.url,
          state: latestDeployment.state,
        });

        // Step 2: Fetch events/logs for this deployment
        const eventsUrl = `https://api.vercel.com/v3/deployments/${deploymentId}/events?direction=backward&limit=100${teamParam}`;
        context.logger.info("Fetching logs for deployment {deploymentId}...", { deploymentId });
        const res = await fetch(eventsUrl, {
          headers: {
            Authorization: `Bearer ${vercelToken}`,
          },
        });

        if (!res.ok) {
          const errText = await res.text();
          throw new Error(`Failed to fetch Vercel deployment events: status ${res.status} - ${errText}`);
        }

        const rawEvents = await res.json();
        const events = Array.isArray(rawEvents) ? rawEvents : [];

        const errorLogs = [];
        for (const log of events) {
          // Dynamic fields parsing since events can have root properties or payload nested
          const text = log.text || log.payload?.text || "";
          const level = log.level || log.payload?.level || log.type || "";
          const lowercaseText = text.toLowerCase();

          const isError = level === "error" || 
                          level === "warning" || 
                          level === "fatal" ||
                          lowercaseText.includes("error") || 
                          lowercaseText.includes("exception");
          if (isError) {
            errorLogs.push({
              id: log.id || log.payload?.id || crypto.randomUUID(),
              text: text,
              timestamp: log.date || log.payload?.date || log.created || Date.now(),
              type: level,
            });
          }
        }

        if (errorLogs.length > 0) {
          context.logger.info("Found {count} errors/warnings in Vercel logs. Sending notification to {email}...", {
            count: errorLogs.length,
            email: notificationEmail,
          });

          // Send notification email via Resend REST API
          const emailBody = `
            <h2>Vercel Production Error Alert</h2>
            <p>Swamp Vercel Monitor found <strong>${errorLogs.length}</strong> warning/error logs in project <strong>${projectId}</strong>:</p>
            <ul>
              ${errorLogs.map(l => `<li><strong>[${new Date(l.timestamp).toISOString()}] [${l.type}]</strong>: ${l.text}</li>`).join("")}
            </ul>
            <br/>
            <p>Sent by Swamp automation.</p>
          `;

          const emailRes = await fetch("https://api.resend.com/emails", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${resendApiKey}`,
            },
            body: JSON.stringify({
              from: "Swamp Monitor <onboarding@resend.dev>",
              to: notificationEmail,
              subject: `Alert: ${errorLogs.length} Vercel Log Issues Found`,
              html: emailBody,
            }),
          });

          if (!emailRes.ok) {
            const errText = await emailRes.text();
            context.logger.error("Failed to send notification email: {error}", { error: errText });
          } else {
            context.logger.info("Notification email sent successfully!");
          }
        } else {
          context.logger.info("No errors or warnings found in the latest Vercel logs.");
        }

        const handle = await context.writeResource("result", "current", {
          checkedAt: new Date().toISOString(),
          totalLogsAnalyzed: events.length,
          errorsFound: errorLogs.length,
          errorLogs: errorLogs,
        });

        return { dataHandles: [handle] };
      },
    },
  },
};
