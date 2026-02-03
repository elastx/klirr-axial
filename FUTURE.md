# Future Plans

## Websocket between frontend and backend
- Real-time frontend updates when new data is synchronized

## Segregate long term and ephemeral synchronization data.
- Current synchronization forces everything to be stored long-term, or the hashes will be wrong.
- Instead, nodes should 

## Chain of trust
- Signing other users so your trustees trust them as well
- Ability to filter everything on "trusted only"
- When I trust someone, That user and everyone they trust is added to my chain of trust.
- As a node owner I can configure my node to only long-term-retain data from users in my chain of trust.

## Code hosting on Axial nodes
- Git server functionality on each Axial node
- Node owner PGP keys used to sign commits
- Chain of trust settings to auto-upgrade when new versions are found on other nodes

## Files
- Receiving nodes set limits on file size.
- UI warns about that limit for files that are over the limit.