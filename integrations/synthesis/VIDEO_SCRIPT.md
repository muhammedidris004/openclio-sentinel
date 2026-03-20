# OpenClio Sentinel Video Script

Use this file as the exact runbook for recording the Synthesis demo video.

Target length:
- 60 to 90 seconds

Goal:
- show the strongest trust proof
- show the strongest cooperation proof
- show the architecture briefly
- close with the public repo

## Windows To Keep Open

Before recording, open these four things:

1. [trust-demo.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/trust-demo.png)
2. [cooperation-demo.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/cooperation-demo.png)
3. [architecture-doc.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/architecture-doc.png)
4. the public repo:
   - `https://github.com/muhammedidris004/openclio-sentinel`

Recommended app layout:
- open the 3 images in Preview
- open the public repo in your browser
- keep only those windows visible
- close chat/debug/setup noise

## Recording Setup

On macOS:

1. Press `Cmd + Shift + 5`
2. Choose `Record Selected Portion`
3. Select the area containing the window you will present
4. Turn microphone on if you want voice narration
5. Start recording

## Exact Order To Show

### Scene 1: Trust Demo

Window:
- [trust-demo.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/trust-demo.png)

Keep it on screen:
- 15 to 20 seconds

What to point at:
- the trust-boundary check
- the blocked or denied action
- the safe fallback inside the workspace

Say this:

```text
This is OpenClio Sentinel, our Synthesis submission.

First, the trust demo. The agent performs an explicit boundary check by trying one harmless inspection outside the allowed workspace. The runtime blocks that access, and the agent immediately falls back to a safe in-workspace inspection. This makes the trust boundary visible and auditable instead of just being a promise.
```

### Scene 2: Cooperation Demo

Window:
- [cooperation-demo.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/cooperation-demo.png)

Keep it on screen:
- 15 to 20 seconds

What to point at:
- delegation
- two subagents
- cooperation evidence counts

Say this:

```text
Second, the cooperation demo. A coordinator delegates a read-only task to two scoped subagents. One inspects trust and reliability risks, and the other proposes the safest next actions for the human. The delegation flow, subagent activity, and completions are exported as evidence.
```

### Scene 3: Architecture

Window:
- [architecture-doc.png](/Users/muhammedidris/Desktop/openclio-enterprise-seed/integrations/synthesis/media/architecture-doc.png)

Keep it on screen:
- 10 to 15 seconds

What to point at:
- runnable Sentinel runtime
- local API
- Synthesis integration exports

Say this:

```text
Architecturally, Sentinel runs on a real OpenClio runtime and local API. The Synthesis integration layer packages the trust and cooperation outputs into evidence bundles that can be reviewed by judges and humans.
```

### Scene 4: Public Repo

Window:
- `https://github.com/muhammedidris004/openclio-sentinel`

Keep it on screen:
- 8 to 12 seconds

What to point at:
- repo name
- public visibility
- docs and runtime files

Say this:

```text
The public submission repo contains the runnable hackathon build, the trust and cooperation demos, and the submission materials. Most agents show capability. OpenClio Sentinel shows capability with bounded authority, auditable behavior, and cooperative workflows a human can trust.
```

## Full Spoken Script

If you want to read it straight through:

```text
This is OpenClio Sentinel, our Synthesis submission.

First, the trust demo. The agent performs an explicit boundary check by trying one harmless inspection outside the allowed workspace. The runtime blocks that access, and the agent immediately falls back to a safe in-workspace inspection. This makes the trust boundary visible and auditable instead of just being a promise.

Second, the cooperation demo. A coordinator delegates a read-only task to two scoped subagents. One inspects trust and reliability risks, and the other proposes the safest next actions for the human. The delegation flow, subagent activity, and completions are exported as evidence.

Architecturally, Sentinel runs on a real OpenClio runtime and local API. The Synthesis integration layer packages the trust and cooperation outputs into evidence bundles that can be reviewed by judges and humans.

The public submission repo contains the runnable hackathon build, the trust and cooperation demos, and the submission materials. Most agents show capability. OpenClio Sentinel shows capability with bounded authority, auditable behavior, and cooperative workflows a human can trust.
```

## Recording Tips

- keep the mouse mostly still
- do not scroll unless necessary
- pause for 1 or 2 seconds before switching windows
- avoid terminal/debug windows in the recording
- do not make the video longer than 90 seconds

## After Recording

1. Stop recording
2. Review once to ensure the text is readable
3. Upload to YouTube as `Unlisted`
4. Save the video URL for the submission API
