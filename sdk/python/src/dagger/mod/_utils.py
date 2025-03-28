import builtins
import dataclasses
import functools
import importlib
import importlib.util
import inspect
import operator
import typing
from collections.abc import Callable, Coroutine
from typing import Any, TypeAlias, TypeVar, cast

import anyio
import anyio.from_thread
import anyio.to_thread
import typing_extensions
from beartype.door import TypeHint, UnionTypeHint
from graphql.pyutils import snake_to_camel

from dagger.mod._arguments import DefaultPath, Ignore, Name
from dagger.mod._types import ContextPath

asyncify = anyio.to_thread.run_sync
syncify = anyio.from_thread.run

T = TypeVar("T")

AwaitableOrValue: TypeAlias = Coroutine[Any, Any, T] | T


async def await_maybe(value: AwaitableOrValue[T]) -> T:
    return await value if inspect.iscoroutine(value) else cast(T, value)


def to_pascal_case(s: str) -> str:
    """Convert a string to PascalCase."""
    return snake_to_camel(s.replace("-", "_"))


def to_camel_case(s: str) -> str:
    """Convert a string to camelCase."""
    return snake_to_camel(s.replace("-", "_"), upper=False)


def normalize_name(name: str) -> str:
    """Remove the last underscore, used to avoid conflicts with reserved words."""
    if name.endswith("_") and name[-2] != "_" and not name.startswith("_"):
        return name.removesuffix("_")
    return name


def get_meta(obj: Any, match: type[T]) -> T | None:
    """Get metadata from an annotated type."""
    if is_initvar(obj):
        return get_meta(obj.type, match)
    if not is_annotated(obj):
        return None
    return next(
        (arg for arg in reversed(typing.get_args(obj)) if isinstance(arg, match)),
        None,
    )


def get_doc(obj: Any) -> str | None:
    """Get the last Doc() in an annotated type or the docstring of an object."""
    if annotated := get_meta(obj, typing_extensions.Doc):
        return annotated.documentation
    if inspect.getmodule(obj) != builtins and (
        inspect.isclass(obj) or inspect.isroutine(obj)
    ):
        doc = inspect.getdoc(obj)
        # By default, a dataclass's __doc__ will be the signature of the class,
        # not None.
        if (
            doc
            and dataclasses.is_dataclass(obj)
            and doc.startswith(f"{obj.__name__}(")
            and doc.endswith(")")
        ):
            doc = None
        return doc
    return None


def get_ignore(obj: Any) -> list[str] | None:
    """Get the last Ignore() of an annotated type."""
    meta = get_meta(obj, Ignore)
    return meta.patterns if meta else None


def get_default_path(obj: Any) -> ContextPath | None:
    """Get the last DefaultPath() of an annotated type."""
    meta = get_meta(obj, DefaultPath)
    return meta.from_context if meta else None


def get_alt_name(annotation: type) -> str | None:
    """Get an alternative name in last Name() of an annotated type."""
    return annotated.name if (annotated := get_meta(annotation, Name)) else None


def is_union(th: TypeHint) -> bool:
    """Returns True if the unsubscripted part of a type is a Union."""
    return isinstance(th, UnionTypeHint)


def is_nullable(th: TypeHint) -> bool:
    """Returns True if the annotation is SomeType | None.

    Does not support Annotated types. Use only on types that have been
    resolved with get_type_hints.
    """
    return th.is_bearable(None)


def non_null(th: TypeHint) -> TypeHint:
    """Removes None from a union.

    Does not support Annotated types. Use only on types that have been
    resolved with get_type_hints.
    """
    if TypeHint(None) not in th:
        return th

    args = (x for x in th.args if x is not type(None))
    return TypeHint(functools.reduce(operator.or_, args))


_T = TypeVar("_T", bound=type)


def is_annotated(annotation: type) -> bool:
    """Check if the given type is an annotated type."""
    return typing.get_origin(annotation) in (
        typing.Annotated,
        typing_extensions.Annotated,
    )


def is_initvar(annotation: type) -> typing.TypeGuard[dataclasses.InitVar]:
    """Check if the given type is a dataclasses.InitVar."""
    return annotation is dataclasses.InitVar or type(annotation) is dataclasses.InitVar


def strip_annotations(t: _T) -> _T:
    """Strip the annotations from a given type."""
    return strip_annotations(typing.get_args(t)[0]) if is_annotated(t) else t


def is_mod_object_type(cls) -> bool:
    """Check if the given class was decorated with @object_type."""
    return hasattr(cls, "__dagger_module__")


def get_alt_constructor(cls: type[T]) -> Callable[..., T] | None:
    """Get classmethod named `create` from object type."""
    if inspect.isclass(cls) and is_mod_object_type(cls):
        fn = getattr(cls, "create", None)
        if inspect.ismethod(fn) and fn.__self__ is cls:
            return fn
    return None


def get_parent_module_doc(obj: type) -> str | None:
    """Get the docstring of the parent module."""
    spec = importlib.util.find_spec(obj.__module__)
    if not spec or not spec.parent:
        return None
    mod = importlib.import_module(spec.parent)
    return inspect.getdoc(mod)
