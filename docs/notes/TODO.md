# TODO's

- Set up custom errors. We'll need them for updates in the repository layer.
- Get recipe functionality finished.
  - Once we have our repo -> service -> controller working, adding the other entities will be easy enough.
- Set up actions - we want simple actions to run on our PR. vet, fmt, build, etc. Dont want to merge something that wont build.
- Unit tests

So far, this feels like its going well. Once we've finished the recipe work, I need to look into how we can spin the backend up in docker. I have a pretty good idea of how it will work, but I need to read some docs.

After that, we can finish the other entities we need. Again, that will be straightforward once we know the recipes are working. I also have ideas for recipe state queries that the front end can hit to fetch everything for a recipe in one request. I think that would be more efficient that mutiple at once, but I need to look into standards to see if that's the correct approach.

When that's all done, we can build the frontend and get this thing deployed. It sounds so easy!
