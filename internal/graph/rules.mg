# prerequisite(Technology, Prerequisite): arrows point toward prerequisites.
Decl technology(T) bound [/name].
Decl prerequisite(T, P) bound [/name, /name].
Decl depends_on(T, P) bound [/name, /name].
Decl impacted_by(P, T) bound [/name, /name].

# Base case and recursive transitive closure (paths of length >= 1).
depends_on(T, P) :- prerequisite(T, P).
depends_on(T, P) :- prerequisite(T, Intermediate), depends_on(Intermediate, P).

# Exclude the source from downstream results, including in cyclic graphs.
impacted_by(P, T) :- depends_on(T, P), T != P.
