# Docker infrastructure inventory predicates.
# These represent observations, not application dependencies.

Decl container(C) bound [/name].
Decl docker_network(N) bound [/name].
Decl connected_to(C, N) bound [/name, /name].
